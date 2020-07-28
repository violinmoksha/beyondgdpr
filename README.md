# BeyondGDPR

BeyondGDPR is based on a very simple, logical-assertion in privacy-ethics regarding end-user data in web and mobile applications.

The assertion: no person, except the end user of an application, should be able to access that user's personal and private data in production. 

This ethical assertion--in a more-perfect use-case than existing implementations across the current Internet--should even exclude the developers being able to access the end-user's production data. Once an application is deployed into real-world production, we should have zero access to that person's protected information (PII, as identified by the GDPR).

GDPR is a great step forward, but it still failed to fully-address this fundamental ethical concern in applications architecture. Letting a government(s)-regulated 3rd party Data Controller have access to user data, although providing an additional layer of protection which excludes developers from access controls, still doesn't directly address the fundamental assertion. The fundamental assertion is: literally only the user should have access to their own PII data, same as the resident of a physical property is the only one with the key to the resident's door. Letting the application developers or 3rd-party data controllers have access to the user's data would then be analogous to allowing the landlord or mortgage company to enter whenever they please. In a more-perfect applications-architecture for data privacy, only the user would have that access, without exception.

BeyondGDPR is a small GoLang-powered middleware proxy-server which directly addresses the ethical concern for any client-server based web or mobile application. Making use of BeyondGDPR in existing systems architecture is fairly simple. An application client--for example, a smartphone app or a web browser--generates a 256-bit key using the publically NSA-vetted AES-GCM cipher, and securely stores it *right there only on the client*, f.e. under the iPhone's device-allocated SecureStorage. The key never leaves that client, except to transit to the BeyondGDPR proxy under SSL (such that the key-in-transit is encrypted under SSL). 

BeyondGDPR never stores that key, but does use it to encrypt/decrypt any plaintext/ciphertext relevant to that user. The web or mobile application only stores ciphertext versions of the user's PII in its DB, therefore no storage of the user's PII exists in plaintext within data views of the application itself--or even within data views of BeyondGDPR. Since the key for decryption lives exclusively on the client--requiring decryption by the non-storing BeyondGDPR middleware exclusively--then the application developers are unable to decrypt the user's PII from any application backend.

As RESTful input, BeyondGDPR takes a JSON with two members from the end-user client under SSL transport: `key` and `ciphertext`. As RESTful output, BeyondGDPR returns a JSON with one member `plaintext` to the end-user client under SSL transport.  

## Initial Load Test Results

Service utilizes multi-thread concurrency and is also Symmetric Multi-Processing (SMP) -enabled (distributes execution across multiple-CPUs where SMP hardware-capability is present), as-native only to GoLang, whereas concurrent JVMs or NodeJS cluster-processes would be much less scalable. Automated volume tests were performed from a cloud in Ashburn, VA directed at a single running instance in Singapore. As we can see, the service withstood 100 VUs--concurrent Virtual Units--over a timeframe of 10seconds, resulting in ~500 r/second total encrypts-decrypts over the network, zero fails (100% 200OK). Thus the concept-implementation in GoLang is proven for real-world application productions.

## TODO

- Kubernetes microservice'ization for additional horizontal-scaling, relative to any app written in GoLang.

## TODO Future-State Consideration

- Separately-keyed assymetric-encrypted storage, and/or additional "backdoor" external-transports under SSL. This would be necessary to answering to legally-warranted government-agencies, iff the use-case were to arise.
