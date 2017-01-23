srv.POST("/v1/users/:user/web/signin", httplib.MakeHandler(srv.signIn))

func (s *APIServer) signIn(w http.ResponseWriter, r *http.Request, p httprouter.Params) (interface{}, error) {
	var req *signInReq
	if err := httplib.ReadJSON(r, &req); err != nil {
		return nil, trace.Wrap(err)
	}
	user := p[0].Value
	sess, err := s.a.SignIn(user, []byte(req.Password))
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return sess, nil
}

	func (a *AuthWithRoles) SignIn(user string, password []byte) (*Session, error) {
		if err := a.permChecker.HasPermission(a.role, ActionSignIn); err != nil {
			return nil, trace.Wrap(err)
		}
		return a.authServer.SignIn(user, password)
	}

		// *** This is where Sign in starts ***
		func (s *AuthServer) SignIn(user string, password []byte) (*Session, error) {
			if err := s.CheckPasswordWOToken(user, password); err != nil {
				return nil, trace.Wrap(err)
			}
			sess, err := s.NewWebSession(user)
			if err != nil {
				return nil, trace.Wrap(err)
			}
			if err := s.UpsertWebSession(user, sess, WebSessionTTL); err != nil {
				return nil, trace.Wrap(err)
			}
			sess.WS.Priv = nil
			return sess, nil
		}

			// CheckPasswordWOToken checks just password without checking HOTP tokens
			// used in case of SSH authentication, when token has been validated
			func (s *IdentityService) CheckPasswordWOToken(user string, password []byte) error {
				if err := services.VerifyPassword(password); err != nil {
					return trace.Wrap(err)
				}
				hash, err := s.GetPasswordHash(user)
				if err != nil {
					return trace.Wrap(err)
				}
				if err = s.IncreaseLoginAttempts(user); err != nil {
					return trace.Wrap(err)
				}
				if err := bcrypt.CompareHashAndPassword(hash, password); err != nil {
					return trace.BadParameter("passwords do not match")
				}
				defer s.ResetLoginAttempts(user)
				return nil
			}

				// VerifyPassword makes sure password satisfies our requirements (relaxed),
				// mostly to avoid putting garbage in
				func VerifyPassword(password []byte) error {
					if len(password) < defaults.MinPasswordLength {
						return trace.BadParameter(
							"password is too short, min length is %v", defaults.MinPasswordLength)
					}
					if len(password) > defaults.MaxPasswordLength {
						return trace.BadParameter(
							"password is too long, max length is %v", defaults.MaxPasswordLength)
					}
					return nil
				}

				// GetPasswordHash returns the password hash for a given user
				func (s *IdentityService) GetPasswordHash(user string) ([]byte, error) {
					hash, err := s.backend.GetVal([]string{"web", "users", user}, "pwd")
					if err != nil {
						if trace.IsNotFound(err) {
							return nil, trace.NotFound("user '%v' is not found", user)
						}
						return nil, trace.Wrap(err)
					}
					return hash, nil
				}

				// IncreaseLoginAttempts bumps "login attempt" counter for the given user. If the counter
				// reaches 'lockAfter' value, it locks the account and returns access denied error.
				func (s *IdentityService) IncreaseLoginAttempts(user string) error {
					bucket := []string{"web", "users", user}

					data, _, err := s.backend.GetValAndTTL(bucket, "lock")
					// unexpected error?
					if err != nil && !trace.IsNotFound(err) {
						return trace.Wrap(err)
					}
					// bump the attempt count
					if len(data) < 1 {
						data = []byte{0}
					}
					// check the attempt count
					if len(data) > 0 && data[0] >= s.lockAfter {
						return trace.AccessDenied("this account has been locked for %v", s.lockDuration)
					}
					newData := []byte{data[0] + 1}
					// "create val" will create a new login attempt counter, or it will
					// do nothing if it's already there.
					//
					// "compare and swap" will bump the counter +1
					s.backend.CreateVal(bucket, "lock", data, s.lockDuration)
					_, err = s.backend.CompareAndSwap(bucket, "lock", newData, s.lockDuration, data)
					return trace.Wrap(err)
				}

				// ResetLoginAttempts resets the "login attempt" counter to zero.
				func (s *IdentityService) ResetLoginAttempts(user string) error {
					err := s.backend.DeleteKey([]string{"web", "users", user}, "lock")
					if trace.IsNotFound(err) {
						return nil
					}
					return trace.Wrap(err)
				}

			func (s *AuthServer) NewWebSession(userName string) (*Session, error) {
				token, err := utils.CryptoRandomHex(TokenLenBytes)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				bearerToken, err := utils.CryptoRandomHex(TokenLenBytes)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				priv, pub, err := s.GetNewKeyPairFromPool()
				if err != nil {
					return nil, err
				}
				ca, err := s.Trust.GetCertAuthority(services.CertAuthID{
					Type:       services.UserCA,
					DomainName: s.DomainName,
				}, true)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				privateKey, err := ca.FirstSigningKey()
				if err != nil {
					return nil, trace.Wrap(err)
				}
				user, err := s.GetUser(userName)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				cert, err := s.Authority.GenerateUserCert(privateKey, pub, user.GetName(), user.GetAllowedLogins(), WebSessionTTL)
				if err != nil {
					return nil, trace.Wrap(err)
				}
				sess := &Session{
					ID:       token,
					Username: user.GetName(),
					WS: services.WebSession{
						Priv:        priv,
						Pub:         cert,
						Expires:     s.clock.Now().UTC().Add(WebSessionTTL),
						BearerToken: bearerToken,
					},
				}
				return sess, nil
			}

			func (s *AuthServer) UpsertWebSession(user string, sess *Session, ttl time.Duration) error {
				return s.Identity.UpsertWebSession(user, sess.ID, sess.WS, ttl)
			}

				// UpsertWebSession updates or inserts a web session for a user and session id
				func (s *IdentityService) UpsertWebSession(user, sid string, session services.WebSession, ttl time.Duration) error {
					bytes, err := json.Marshal(session)
					if err != nil {
						return trace.Wrap(err)
					}
					err = s.backend.UpsertVal([]string{"web", "users", user, "sessions"},
						sid, bytes, ttl)
					if trace.IsNotFound(err) {
						return trace.NotFound("user '%v' is not found", user)
					}
					return trace.Wrap(err)
				}
