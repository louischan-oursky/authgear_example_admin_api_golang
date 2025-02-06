package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// TODO: Replace with your project ID here.
// It is the first part of your Authgear endpoint.
// e.g. The project ID is "myapp" for "https://myapp.authgear.cloud"
const ProjectID = "accounts"

// TODO: Replace with the key ID you see in the portal.
const KeyID = "8fce2c32-0a96-4431-ba40-a7f068eed276"

// TODO: Follow the below guide and obtain the private key PEM file, and place it next to main.go.
// https://docs.authgear.com/reference/apis/admin-api/authentication-and-security#obtaining-the-private-key-for-signing-jwt
const PrivateKeyPath = "./admin-api-private-key.pem"

// TODO: Replace with your project endpoint.
// e.g. "https://myapp.authgear.cloud"
const AuthgearEndpoint = "http://localhost:3100"

func prepareToken() (token string, err error) {
	// On production, you want to read and parse it once only.
	f, err := os.Open(PrivateKeyPath)
	if err != nil {
		return
	}
	defer f.Close()
	jwkSet, err := jwk.ParseReader(f, jwk.WithPEM(true))
	if err != nil {
		return
	}

	key, _ := jwkSet.Key(0)
	key.Set("kid", KeyID)

	now := time.Now().UTC()
	payload := jwt.New()
	_ = payload.Set(jwt.AudienceKey, ProjectID)
	_ = payload.Set(jwt.IssuedAtKey, now.Unix())
	// The token will expire in 5 minutes.
	_ = payload.Set(jwt.ExpirationKey, now.Add(5*time.Minute).Unix())

	// The alg MUST be RS256.
	alg := jwa.RS256
	hdr := jws.NewHeaders()
	hdr.Set("typ", "JWT")

	buf, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	tokenBytes, err := jws.Sign(buf, jws.WithKey(alg, key, jws.WithProtectedHeaders(hdr)))
	if err != nil {
		return
	}

	token = string(tokenBytes)
	return
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLError any

type GraphQLResponse struct {
	Data   any            `json:"data,omitempty"`
	Errors []GraphQLError `json:"errors,omitempty"`
}

func (r *GraphQLResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r *GraphQLResponse) Error() string {
	if r.HasErrors() {
		b, err := json.Marshal(r)
		if err != nil {
			panic(err)
		}
		return string(b)
	}
	return ""
}

func performGraphQLRequest(ctx context.Context, httpClient *http.Client, token string, request GraphQLRequest) (*GraphQLResponse, error) {
	u, err := url.JoinPath(AuthgearEndpoint, "/_api/admin/graphql")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	body := bytes.NewReader(bodyBytes)
	r, err := http.NewRequestWithContext(ctx, "POST", u, body)
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var graphqlResponse GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&graphqlResponse)
	if err != nil {
		return nil, err
	}

	if graphqlResponse.HasErrors() {
		return nil, &graphqlResponse
	}
	return &graphqlResponse, nil
}

func performGraphQLRequestAndPrintToStdout(ctx context.Context, token string, request GraphQLRequest) error {
	resp, err := performGraphQLRequest(ctx, http.DefaultClient, token, request)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", string(b))
	return nil
}

func main() {
	token, err := prepareToken()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// This showcases how to search users with email ending with "oursky.com".
	err = performGraphQLRequestAndPrintToStdout(ctx, token, GraphQLRequest{
		Query: `query searchUserExample($first: Int!, $keyword: String!) {
			users(first: $first, searchKeyword: $keyword) {
				edges {
					node {
						id
						standardAttributes
					}
				}
			}
		}`,
		OperationName: "searchUserExample",
		Variables: map[string]interface{}{
			"first":   5,
			"keyword": "example.com",
		},
	})
	if err != nil {
		panic(err)
	}

	// This showcases how to create a user with a email.
	err = performGraphQLRequestAndPrintToStdout(ctx, token, GraphQLRequest{
		Query: `mutation createUserExample($email: String!) {
			createUser(
				input: {definition: {loginID: {key: "email", value: $email } } }
			) {
				user {
					id
					standardAttributes
				}
			}
		}
		`,
		OperationName: "createUserExample",
		Variables: map[string]interface{}{
			"email": "user@example.com",
		},
	})
	if err != nil {
		panic(err)
	}
}
