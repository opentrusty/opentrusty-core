// Copyright 2026 The OpenTrusty Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/opentrusty/opentrusty-core/audit"
	"github.com/opentrusty/opentrusty-core/id"
)

// Service provides OAuth2 client management business logic.
//
// Purpose: Implementation of client registration, validation, and lifecycle rules.
// Domain: OAuth2
type Service struct {
	clientRepo  ClientRepository
	auditLogger audit.Logger
}

// NewService creates a new client management service.
//
// Purpose: Constructor for the client management service.
// Domain: OAuth2
// Audited: No
// Errors: None
func NewService(clientRepo ClientRepository, auditLogger audit.Logger) *Service {
	return &Service{
		clientRepo:  clientRepo,
		auditLogger: auditLogger,
	}
}

// RegisterClient validates and creates a new OAuth2 client.
//
// Purpose: Enforces system rules on new client registrations and persists them.
// Domain: OAuth2
// Audited: Yes (ClientCreated)
// Errors: ErrInvalidClientURI, ErrInvalidRedirectURI, System errors
func (s *Service) RegisterClient(ctx context.Context, tenantID, userID string, c *Client) (*Client, error) {
	if err := s.validateClient(c); err != nil {
		return nil, err
	}

	if c.ID == "" {
		c.ID = id.NewUUIDv7()
	}
	if c.ClientID == "" {
		c.ClientID = id.NewUUIDv7()
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	c.UpdatedAt = time.Now()

	if err := s.clientRepo.Create(ctx, c); err != nil {
		return nil, err
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeClientCreated,
		TenantID:   tenantID,
		ActorID:    userID,
		Resource:   audit.ResourceClient,
		TargetName: c.ClientName,
		TargetID:   c.ClientID,
		Metadata: map[string]any{
			"client_id":   c.ClientID,
			"client_name": c.ClientName,
		},
	})

	return c, nil
}

// ListClients retrieves all OAuth2 clients for a tenant
func (s *Service) ListClients(ctx context.Context, tenantID string) ([]*Client, error) {
	return s.clientRepo.ListByTenant(ctx, tenantID)
}

// GetClient retrieves an OAuth2 client by internal ID
func (s *Service) GetClient(ctx context.Context, tenantID, id string) (*Client, error) {
	return s.clientRepo.GetByID(ctx, tenantID, id)
}

// GetClientByClientID retrieves an OAuth2 client by external client_id
func (s *Service) GetClientByClientID(ctx context.Context, tenantID, clientID string) (*Client, error) {
	return s.clientRepo.GetByClientID(ctx, tenantID, clientID)
}

// DeleteClient deletes an OAuth2 client
func (s *Service) DeleteClient(ctx context.Context, tenantID, id string, actorID string) error {
	c, err := s.clientRepo.GetByID(ctx, tenantID, id)
	if err != nil {
		return err
	}

	if err := s.clientRepo.Delete(ctx, tenantID, id); err != nil {
		return err
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeClientDeleted,
		TenantID:   tenantID,
		ActorID:    actorID,
		Resource:   audit.ResourceClient,
		TargetName: c.ClientName,
		TargetID:   c.ClientID,
		Metadata: map[string]any{
			"client_id": c.ClientID,
		},
	})
	return nil
}

// UpdateClient updates an existing OAuth2 client
func (s *Service) UpdateClient(ctx context.Context, c *Client, actorID string) error {
	if err := s.validateClient(c); err != nil {
		return err
	}
	c.UpdatedAt = time.Now()
	if err := s.clientRepo.Update(ctx, c); err != nil {
		return err
	}

	s.auditLogger.Log(ctx, audit.Event{
		Type:       audit.TypeClientUpdated,
		TenantID:   c.TenantID,
		ActorID:    actorID,
		Resource:   audit.ResourceClient,
		TargetName: c.ClientName,
		TargetID:   c.ClientID,
		Metadata: map[string]any{
			"client_id": c.ClientID,
		},
	})
	return nil
}

func (s *Service) validateClient(c *Client) error {
	if c.ClientURI != "" {
		if _, err := url.ParseRequestURI(c.ClientURI); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidClientURI, err)
		}
	}

	for _, uri := range c.RedirectURIs {
		if _, err := url.ParseRequestURI(uri); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidRedirectURI, uri)
		}
	}
	return nil
}
