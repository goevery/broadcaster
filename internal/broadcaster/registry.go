package broadcaster

import (
	"errors"
	"sync"

	"go.uber.org/zap"
)

type Registry interface {
	Broadcast(message Message)
	Register(channelId string, connection Connection) error
	Unregister(channelId string, connectionId string)
	Disconnect(connectionId string)
}

type InMemoryRegistry struct {
	logger *zap.Logger
	mu     sync.RWMutex

	connections          map[string]Connection
	connectionsByChannel map[string]map[string]struct{}
	channelsByConnection map[string]map[string]struct{}
}

func NewInMemoryRegistry(
	logger *zap.Logger,
) *InMemoryRegistry {
	return &InMemoryRegistry{
		logger:               logger,
		connections:          make(map[string]Connection),
		connectionsByChannel: make(map[string]map[string]struct{}),
		channelsByConnection: make(map[string]map[string]struct{}),
	}
}

func (r *InMemoryRegistry) Broadcast(message Message) {
	r.mu.RLock()

	connectionIds, ok := r.connectionsByChannel[message.ChannelId]
	if !ok {
		r.mu.RUnlock()

		return
	}

	connections := make([]Connection, 0, len(connectionIds))
	for connectionId := range connectionIds {
		if connection, ok := r.connections[connectionId]; ok {
			connections = append(connections, connection)
		}
	}

	var staleConnectionIds []string

	for _, connection := range connections {
		select {
		case connection.Send <- message:
		default:
			r.logger.Warn("connection send channel is full, closing connection",
				zap.String("connectionId", connection.Id))

			staleConnectionIds = append(staleConnectionIds, connection.Id)
		}
	}

	r.mu.RUnlock()

	if len(staleConnectionIds) == 0 {
		return
	}

	r.mu.Lock()

	for _, connectionId := range staleConnectionIds {
		r.disconnectLocked(connectionId)
	}

	r.mu.Unlock()
}

func (r *InMemoryRegistry) Register(channelId string, connection Connection) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure map for the channel exists
	if _, ok := r.connectionsByChannel[channelId]; !ok {
		r.connectionsByChannel[channelId] = make(map[string]struct{})
	}

	// Check if connection is already registered to the channel
	if _, ok := r.connectionsByChannel[channelId][connection.Id]; ok {
		return errors.New("connection already registered to channel")
	}

	r.connectionsByChannel[channelId][connection.Id] = struct{}{}
	r.connections[connection.Id] = connection

	if _, ok := r.channelsByConnection[connection.Id]; !ok {
		r.channelsByConnection[connection.Id] = make(map[string]struct{})
	}

	r.channelsByConnection[connection.Id][channelId] = struct{}{}

	return nil
}

func (r *InMemoryRegistry) Unregister(channelId string, connectionId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	connectionChannels, ok := r.channelsByConnection[connectionId]
	if !ok {
		return
	}

	delete(connectionChannels, channelId)
	if len(connectionChannels) == 0 {
		delete(r.channelsByConnection, connectionId)
		delete(r.connections, connectionId)
	}

	channelConnections, ok := r.connectionsByChannel[channelId]
	if !ok {
		panic("inconsistent state: channel not found in connectionsByChannel")
	}

	delete(channelConnections, connectionId)
	if len(channelConnections) == 0 {
		delete(r.connectionsByChannel, channelId)
	}
}

func (r *InMemoryRegistry) Disconnect(connectionId string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.disconnectLocked(connectionId)
}

// IMPORTANT: It must be called only when a write lock is already held.
func (r *InMemoryRegistry) disconnectLocked(connectionId string) {
	connection, ok := r.connections[connectionId]
	if !ok {
		return
	}

	connectionChannels, ok := r.channelsByConnection[connectionId]
	if !ok {
		panic("inconsistent state: connection not found in channelsByConnection")
	}

	for channelId := range connectionChannels {
		channelConnections, ok := r.connectionsByChannel[channelId]
		if !ok {
			panic("inconsistent state: channel not found in connectionsByChannel")
		}

		delete(channelConnections, connectionId)
		if len(channelConnections) == 0 {
			delete(r.connectionsByChannel, channelId)
		}
	}

	delete(r.channelsByConnection, connectionId)
	delete(r.connections, connectionId)
	close(connection.Send)
}
