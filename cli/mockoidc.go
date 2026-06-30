package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/oauth2-proxy/mockoidc"
	"github.com/spf13/cobra"
)

const (
	defaultAccessTTL  = 2 * time.Minute
	defaultRefreshTTL = 60 * time.Minute
)

var mockOidcCmd = &cobra.Command{
	Use:    "mockoidc",
	Short:  "Runs a mock OIDC server for testing",
	Long:   "This internal command runs a mock OpenID Connect server for testing purposes.",
	Hidden: true,
	RunE:   runMockOIDC,
}

func runMockOIDC(cmd *cobra.Command, args []string) error {
	clientID := os.Getenv("MOCKOIDC_CLIENT_ID")
	if clientID == "" {
		return fmt.Errorf("MOCKOIDC_CLIENT_ID not defined")
	}

	clientSecret := os.Getenv("MOCKOIDC_CLIENT_SECRET")
	if clientSecret == "" {
		return fmt.Errorf("MOCKOIDC_CLIENT_SECRET not defined")
	}

	addr := os.Getenv("MOCKOIDC_ADDR")
	if addr == "" {
		addr = "localhost"
	}

	portStr := os.Getenv("MOCKOIDC_PORT")
	if portStr == "" {
		return fmt.Errorf("MOCKOIDC_PORT not defined")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid MOCKOIDC_PORT: %w", err)
	}

	// Optional issuer override for when binding to 0.0.0.0
	issuerOverride := os.Getenv("MOCKOIDC_ISSUER")

	accessTTL := defaultAccessTTL
	if ttlStr := os.Getenv("MOCKOIDC_ACCESS_TTL"); ttlStr != "" {
		accessTTL, err = time.ParseDuration(ttlStr)
		if err != nil {
			return fmt.Errorf("invalid MOCKOIDC_ACCESS_TTL: %w", err)
		}
	}

	userStr := os.Getenv("MOCKOIDC_USERS")
	if userStr == "" {
		return fmt.Errorf("MOCKOIDC_USERS not defined")
	}

	var users []mockoidc.MockUser
	if err := json.Unmarshal([]byte(userStr), &users); err != nil {
		return fmt.Errorf("unmarshalling MOCKOIDC_USERS: %w", err)
	}

	fmt.Printf("Mock OIDC: Loading %d users\n", len(users))
	fmt.Printf("Mock OIDC: Access token TTL: %s\n", accessTTL)

	mock, err := createMockOIDC(clientID, clientSecret, accessTTL, users)
	if err != nil {
		return fmt.Errorf("creating mock OIDC: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return fmt.Errorf("listening on %s:%d: %w", addr, port, err)
	}

	// If issuer override is set, we need to wrap the handler to rewrite URLs
	if issuerOverride != "" {
		fmt.Printf("Mock OIDC: Using issuer override: %s\n", issuerOverride)
		if err := mock.Start(listener, nil); err != nil {
			return fmt.Errorf("starting mock OIDC: %w", err)
		}
		// Get the original issuer to replace in responses
		originalIssuer := mock.Issuer()
		mock.Shutdown()

		// Create a new mock and serve with URL rewriting
		mock, err = createMockOIDC(clientID, clientSecret, accessTTL, users)
		if err != nil {
			return fmt.Errorf("creating mock OIDC: %w", err)
		}

		listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", addr, port))
		if err != nil {
			return fmt.Errorf("listening on %s:%d: %w", addr, port, err)
		}

		// Add middleware to rewrite issuer in responses
		mock.AddMiddleware(func(h http.Handler) http.Handler {
			return &issuerRewriter{
				handler:        h,
				originalIssuer: originalIssuer,
				newIssuer:      issuerOverride,
			}
		})

		if err := mock.Start(listener, nil); err != nil {
			return fmt.Errorf("starting mock OIDC: %w", err)
		}

		fmt.Printf("Mock OIDC server listening on %s\n", listener.Addr().String())
		fmt.Printf("Issuer (rewritten): %s\n", issuerOverride)
	} else {
		if err := mock.Start(listener, nil); err != nil {
			return fmt.Errorf("starting mock OIDC: %w", err)
		}
		fmt.Printf("Mock OIDC server listening on %s\n", listener.Addr().String())
		fmt.Printf("Issuer: %s\n", mock.Issuer())
	}

	// Block forever
	select {}
}

func createMockOIDC(clientID, clientSecret string, accessTTL time.Duration, users []mockoidc.MockUser) (*mockoidc.MockOIDC, error) {
	keypair, err := mockoidc.NewKeypair(nil)
	if err != nil {
		return nil, err
	}

	userQueue := mockoidc.UserQueue{}
	for _, user := range users {
		userQueue.Push(&user)
	}

	mock := mockoidc.MockOIDC{
		ClientID:                      clientID,
		ClientSecret:                  clientSecret,
		AccessTTL:                     accessTTL,
		RefreshTTL:                    defaultRefreshTTL,
		CodeChallengeMethodsSupported: []string{"plain", "S256"},
		Keypair:                       keypair,
		SessionStore:                  mockoidc.NewSessionStore(),
		UserQueue:                     &userQueue,
		ErrorQueue:                    &mockoidc.ErrorQueue{},
	}

	mock.AddMiddleware(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("Mock OIDC: %s %s\n", r.Method, r.URL.Path)
			h.ServeHTTP(w, r)
		})
	})

	return &mock, nil
}

// issuerRewriter wraps an http.Handler and rewrites the issuer URL in responses
type issuerRewriter struct {
	handler        http.Handler
	originalIssuer string
	newIssuer      string
}

func (ir *issuerRewriter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Wrap the response writer to capture and modify the response
	wrapper := &responseWrapper{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		statusCode:     http.StatusOK,
	}
	ir.handler.ServeHTTP(wrapper, r)

	// Rewrite issuer in the response body
	body := wrapper.body.String()
	body = strings.ReplaceAll(body, ir.originalIssuer, ir.newIssuer)

	// Write the modified response
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(wrapper.statusCode)
	w.Write([]byte(body))
}

type responseWrapper struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}
