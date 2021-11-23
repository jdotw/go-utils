# Banking as a Service

This repository contains the common code used by BaaS API clusters and
microservices.

## Decisions

### Tenets

- Bias for performance and simplicity

### Tech Stack

- Backend is written in Go
- PostgreSQL is our relational database
- MongoDB is our noSQL database
- Code is run from Docker containers
- HTTP and/or gRPC as transport between services and also with clients

### Architecture

- Following BIAN Service Domains, but writing our own API specs for those domains
- Each of the Service Domain's (ref BIAN) is a self-contained microservice
  that is its own source of truth. It may be integrating with an underlying
  system of record (e.g. SAP) but the service assumes it is the source of
  truth, maintains its own data and consistency, with the underlying system
  as a fallback / backup.
- No experience layer. Just the back-end, and the front-end.
- We create the backend and the client SDK for end-to-end ownership
- Client SDK would present a native (Swift & Kotlin) dev-friendly API to Customer, Accounts,
  Transactions, Product and Payments. Back-end is built for modular service
  domains following the BIAN 9.1 reference arch.
- Complexity in the back-end service domain interactions is handled in the client SDK libraries, not via additional tiers of services.
