package server

import (
	"net/http"
	"net/url"

	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/utils/sentry"
	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// JWKS returns the JSON web key set for the public RSA keys that are currently being
// used by Epistolary to sign JWT acccess and refresh tokens. External callers can use
// these keys to verify that a JWT token was in fact issued by the Epistolary API.
func (s *Server) JWKS(c *gin.Context) {
	jwks := jwk.NewSet()
	for keyid, pubkey := range s.tokens.Keys() {
		key, err := jwk.FromRaw(pubkey)
		if err != nil {
			sentry.Error(c).Err(err).Str("kid", keyid.String()).Msg("could not parse tokens public key")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("an internal error occurred"))
			return
		}

		if err = key.Set(jwk.KeyIDKey, keyid.String()); err != nil {
			sentry.Error(c).Err(err).Str("kid", keyid.String()).Msg("could not set tokens public key id")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("an internal error occurred"))
			return
		}

		if err = key.Set(jwk.KeyUsageKey, jwk.ForSignature); err != nil {
			sentry.Error(c).Err(err).Str("kid", keyid.String()).Msg("could not set tokens public key use")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("an internal error occurred"))
			return
		}

		// NOTE: the algorithm should match the signing method in tokens.go
		if err = key.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
			sentry.Error(c).Err(err).Str("kid", keyid.String()).Msg("could not set tokens public key algorithm")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("an internal error occurred"))
			return
		}

		if err = jwks.AddKey(key); err != nil {
			sentry.Error(c).Err(err).Str("kid", keyid.String()).Msg("could not add key to jwks")
			c.JSON(http.StatusInternalServerError, api.ErrorResponse("an internal error occurred"))
			return
		}
	}
	c.JSON(http.StatusOK, jwks)
}

// Returns a JSON document with the OpenID configuration as defined by the OpenID
// Connect standard: https://connect2id.com/learn/openid-connect. This document helps
// clients understand how to authenticate with Epistolary.
// TODO: once OpenID endpoints have been configured add them to this JSON response
func (s *Server) OpenIDConfiguration(c *gin.Context) {
	// Parse the token issuer for the OpenID configuration
	base, err := url.Parse(s.conf.Token.Issuer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse("openid is not configured correctly"))
		return
	}

	openid := &api.OpenIDConfiguration{
		Issuer:                        base.ResolveReference(&url.URL{Path: "/"}).String(),
		JWKSURI:                       base.ResolveReference(&url.URL{Path: "/.well-known/jwks.json"}).String(),
		ScopesSupported:               []string{"openid", "profile", "email"},
		ResponseTypesSupported:        []string{"token", "id_token"},
		CodeChallengeMethodsSupported: []string{"S256", "plain"},
		ResponseModesSupported:        []string{"query", "fragment", "form_post"},
		SubjectTypesSupported:         []string{"public"},
		IDTokenSigningAlgValues:       []string{"HS256", "RS256"},
		TokenEndpointAuthMethods:      []string{"client_secret_basic", "client_secret_post"},
		ClaimsSupported:               []string{"aud", "name", "email", "permissions", "exp", "iat", "iss", "sub"},
		RequestURIParameterSupported:  false,
	}

	c.JSON(http.StatusOK, openid)
}

// Writes the security.txt file generated from https://securitytxt.org/ and digitally
// signed with the info@rotational.io PGP keys to alert security researchers to our
// security policies and allow them to contact us with any security flaws.
func (s *Server) SecurityTxt(c *gin.Context) {
	c.String(http.StatusOK, securityTxt)
}

const securityTxt = `-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

Contact: mailto:server@bengfort.com
Expires: 2026-04-07T12:00:00.000Z
Encryption: https://keys.openpgp.org/vks/v1/by-fingerprint/78A2DCBFA18B6D649A889AE76737E124F0D6AB12
Preferred-Languages: en
Canonical: https://epistolary.app/.well-known/security.txt
-----BEGIN PGP SIGNATURE-----

iQIzBAEBCAAdFiEEeKLcv6GLbWSaiJrnZzfhJPDWqxIFAmNoLBUACgkQZzfhJPDW
qxIZdxAAukxFtD1JNofcyk9QGaCWWeZmZEmv+h6rSQ6Y+TRCVLl5HeUZG0LlE84g
/YCTlkfqQ96HYUZHyVt1DDwOmaIHVZfT+foGd8Ng0d//uhhzAuPGK2lQ+wfCNTBf
yvjbm0S8+YO9oPPAJlELN8ARZ6+aS5fUE9ywZ5QUb67V+LPCgvNjiHxVyJWN7lQA
67Hr2ACqUt2Hivqso9J9QvNsnwGO/KPTkcLzWonFYSKfGtTxwuLmfqxqFMpmwx9I
CJG5PBgV5nMEOEypWEwoDIjzHce1hOcEbABpTPeXZKQat18Na1w0oxhBS6DJErBa
j/zqydoYNMlb3tiYtYOvh45ESAYPguk6NTPVaCyGPBaNEv86ldJb1IVZc4vv1L/u
nPtYRWSt+kdkkYKMsNWl6TnHRvwffts0xdHLueDyFMRG2ceCGS3Z3Gj4qgqH+NLR
s6It45C2LoQxwaoJ1Y3jhuJt1K3OuAfh8zalAxEm8p1PeBvSVNPK5zpg3SpWaD0a
Q2ClR/8kA/iBhCOGsupSijirrpohyu0GpLFLB6klNH5bBnZxEVoAaV8gmOayUmUY
8DELToe3IbFSRzyt0sdlt5V5w8PuNjhMrNuFUffcxg3EcIoED08FiPXB+GPpPJJL
1/17Remr0t6NwiycKf67BY9/a4fTuG7xE4wV7BdgeX5zZ9m/fxw=
=ZhqB
-----END PGP SIGNATURE-----
`
