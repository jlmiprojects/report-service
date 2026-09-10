package repository

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultConnectionName = "default"

// MongoConnections is the registry of named MongoDB datasources a `mongo`
// report DataAction can query against. Connections are defined under
// `mongo_connections` in application.json and opened lazily on first use, so an
// unreachable optional datasource only fails the report that needs it rather
// than blocking service start-up.
type MongoConnections struct {
	mu      sync.Mutex
	defs    map[string]utils.MongoConnection
	clients map[string]*mongo.Client
}

// NewMongoConnections builds the registry from config. It always guarantees a
// "default" entry: the configured `mongo_connections.default` when present,
// otherwise one synthesized from the primary `mongo` block.
func NewMongoConnections(config *utils.Config) *MongoConnections {
	defs := make(map[string]utils.MongoConnection)

	for name, def := range config.MongoConnections {
		if def == nil {
			continue
		}
		defs[name] = *def
	}

	if _, ok := defs[defaultConnectionName]; !ok {
		defs[defaultConnectionName] = utils.MongoConnection{
			Uri:            config.Mongo.Uri,
			Database:       config.Mongo.Database,
			ConnectTimeout: config.Mongo.ConnectTimeout,
			ReadTimeout:    config.Mongo.ReadTimeout,
		}
	}

	return &MongoConnections{
		defs:    defs,
		clients: make(map[string]*mongo.Client),
	}
}

// Names returns the configured connection names (for logging / diagnostics).
func (m *MongoConnections) Names() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.defs))
	for name := range m.defs {
		names = append(names, name)
	}
	return names
}

// Connection returns the driver client for a named connection (blank =>
// "default"), connecting and pinging it on first use and caching it
// thereafter. Exported for reportapi's MongoHandle.
func (m *MongoConnections) Connection(name string) (*mongo.Client, utils.MongoConnection, error) {
	return m.connection(name)
}

// connection returns the driver client for a named connection, connecting and
// pinging it on first use and caching it thereafter.
func (m *MongoConnections) connection(name string) (*mongo.Client, utils.MongoConnection, error) {
	if name == "" {
		name = defaultConnectionName
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	def, ok := m.defs[name]
	if !ok {
		return nil, utils.MongoConnection{}, fmt.Errorf("unknown mongo connection %q", name)
	}

	if client, ok := m.clients[name]; ok {
		return client, def, nil
	}

	if def.Uri == "" {
		return nil, def, fmt.Errorf("mongo connection %q has no uri configured", name)
	}

	connectTimeout := def.ConnectTimeout
	if connectTimeout <= 0 {
		connectTimeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	bsonOpts := &options.BSONOptions{UseJSONStructTags: true, NilSliceAsEmpty: true}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(def.Uri).SetServerAPIOptions(serverAPI).SetBSONOptions(bsonOpts)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, def, fmt.Errorf("connect mongo connection %q: %w", name, err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, def, fmt.Errorf("ping mongo connection %q: %w", name, err)
	}

	slog.Info("opened mongo report connection", "name", name, "database", def.Database)
	m.clients[name] = client
	return client, def, nil
}

// Close disconnects every opened connection.
func (m *MongoConnections) Close(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, client := range m.clients {
		if err := client.Disconnect(ctx); err != nil {
			slog.Warn("failed to disconnect mongo report connection", "name", name, "error", err)
		}
		delete(m.clients, name)
	}
}
