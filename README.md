# BeyondGDPR

BeyondGDPR is based on a very simple, logical-assertion in privacy-ethics regarding end-user data in web and mobile applications.

The assertion: no person, except the end user of an application, should be able to access that user's personal and private data in production. 

This ethical assertion--in a more-perfect use-case than existing implementations across the current Internet--should even exclude the developers being able to access the end-user's production data. Once an application is deployed into real-world production, we should have zero access to that person's protected information (PII, as identified by the GDPR).

GDPR is great step forward, but it still failed to fully-address this fundamental ethical concern in applications architecture. Letting a government(s)-regulated 3rd party Data Controller have access to user data, although providing an additional layer of protection which excludes developers from access controls, still doesn't directly address the fundamental assertion. The fundamental assertion is: literally only the user should have access to their own PII data, same as the resident of a physical property is the only one with the key to the resident's door. Letting the application developers or 3rd-party data controllers have access to the user's data would then be analogous to allowing the landlord or mortgage company to enter whenever they please. In a more-perfect applications-architecture for data privacy, only the user would have that access, without exception.

BeyondGDPR is a small GoLang-powered middleare proxy-server which directly addresses the ethical concern for any client-server based web or mobile application. Making use of BeyondGDPR in existing systems architecture is fairly simple. An application client--for example, a smartphone app or a web browser--generates a 256-bit key using the publically NSA-vetted AES-GCM cipher, and securely stores it *right there only on the client*, f.e. under the smartphone's device-provisioned Secure Storage. The key never leaves that client, except to transit to the BeyondGDPR proxy under SSL (such that the key-in-transit is encrypted under SSL). BeyondGDPR never stores that key, but does use it to encrypt/decrypt any plaintext/ciphertext relevant to that user. The web or mobile application only stores ciphertext versions of the user's PII in its DB, therefore no storage of the user's PII exists in plaintext within data views of the application itself. Since the key for decryption lives exclusively on the client--requiring decryption by the non-storing BeyondGDPR middleware, exclusively--then the application developers are unable to decrypt the user's PII from any application backend.

As RESTful input, BeyondGDPR takes a JSON with two members from the end-user client under SSL transport: key and ciphertext. As RESTful output, BeyondGDPR returns a plaintext to the end-user client under SSL transport.  

## Initial Load Test Results

Service utilizes multi-thread concurrency native and also Symmetric Multi-Processing aka SMP (multi-CPU cores), as virtues native to GoLang (whereas JVMs or NodeJS cluster-instances would serve as less-powerful, dependent on essentially one or the other versus the ability to scale both vertically and horizontally). Automated volume tests were performed from a cloud in Ashburn, VA directed at a single running instance in Singapore. As we can see, the service withstood 100 VUs over 10seconds resulting in ~500 r/second of concurrent encrypt-decrypt over the network, with zero fails (100% 200OK). Thus the implementation exclusively in GoLang-only proves the concept-feasibility for real-world application architectures.

## TODO

- Kubernetes microservice'ization for additional horizontal-scaling relative to app architectures which have been written in GoLang-only. ;-)

## TODO Considerations

- Possible separately-keyed assymetric-encrypted storage, or additional external-transport under SSL for warranted government-agency operations. 8-)
