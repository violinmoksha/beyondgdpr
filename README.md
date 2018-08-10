# BeyondGDPR

BeyondGDPR is based on a very simple ethical assertion.

That assertion is: no one, except the user of an application, should be able to access that user's personal and private data. This ethical assertion should apply to the developers as well, since we too should have zero access to a real-world user's personal information (PII, as identified by the GDPR).

GDPR is great, but it still failed (at the time of this writing) to fully-address this fundamental ethical concern, that only the user should have access to the user's data. Letting say, a government(s)-regulated 3rd party Data Controller have access to that data--instead of the application developers--although providing an additional layer of protection, still doesn't directly address the fundamental assertion, which is that only the user should have access to that user's PII.

BeyondGDPR is a small GoLang-powered server which directly addresses the ethical concern for any application. Making use of BeyondGDPR in existing systems architecture is fairly simple. An application client or web browser generates a 256-bit key for AES-GCM and securely stores it *only on the client*, f.e. on a mobile phone or in a web browser. The key never leaves that client, except to transit to BeyondGDPR under SSL. BeyondGDPR never stores said key, but does use it to encrypt/decrypt any plaintext/ciphertext relevant to that user. The application goes on storing only the resultant encrypted data returned by BeyondGDPR (and never plaintexts), into its own pre-existing backend data storage. Since the key lives exclusively on the client, application developers are unable to decrypt from application backend.

## TODO

- Load testing.
- Go Concurrency for vertical scaling.
- Kubernetes for horizontal scaling.
