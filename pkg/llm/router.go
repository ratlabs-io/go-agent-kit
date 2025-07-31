package llm

import (
	"context"
	"fmt"
	"strings"
)

// RouterClient implements the Client interface and routes requests to registered providers.
type RouterClient struct {
	clients map[string]Client
}

// NewRouterClient creates a new router client.
func NewRouterClient() *RouterClient {
	return &RouterClient{
		clients: make(map[string]Client),
	}
}

// Register registers a client for a specific provider.
func (r *RouterClient) Register(provider string, client Client) {
	r.clients[strings.ToLower(provider)] = client
}

// Complete implements Client.Complete by routing to the appropriate provider based on the model.
func (r *RouterClient) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	// Parse provider and model from the model string (format: "provider/model")
	provider, model, err := r.parseProviderAndModel(req.Model)
	if err != nil {
		return nil, err
	}

	// Get client for provider
	client := r.clients[provider]

	// Update the request with the actual model name (without provider prefix)
	req.Model = model

	// Route to appropriate client
	response, err := client.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Close implements Client.Close by closing all registered clients.
func (r *RouterClient) Close() error {
	for _, client := range r.clients {
		if err := client.Close(); err != nil {
			return err
		}
	}
	return nil
}

// parseProviderAndModel parses a model string in the format "provider/model".
// Returns an error if the format is invalid or the provider is not registered.
func (r *RouterClient) parseProviderAndModel(modelString string) (provider, model string, err error) {
	// Check if the model string contains a provider prefix
	if !strings.Contains(modelString, "/") {
		return "", "", fmt.Errorf("model string must be in format 'provider/model', got: %s", modelString)
	}

	parts := strings.SplitN(modelString, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model format, expected 'provider/model', got: %s", modelString)
	}

	provider = strings.ToLower(parts[0])
	model = parts[1]

	// Validate that the provider is registered
	if _, exists := r.clients[provider]; !exists {
		return "", "", fmt.Errorf("provider %s is not registered", provider)
	}

	return provider, model, nil
}

// GetRegisteredProviders returns a list of registered providers.
func (r *RouterClient) GetRegisteredProviders() []string {
	var providers []string
	for provider := range r.clients {
		providers = append(providers, provider)
	}
	return providers
}

// IsProviderRegistered checks if a provider is registered.
func (r *RouterClient) IsProviderRegistered(provider string) bool {
	_, exists := r.clients[strings.ToLower(provider)]
	return exists
}