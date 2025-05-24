# HTTP Server – Boot.dev Course Project

This project is a simple HTTP server built as part of the [Boot.dev](https://boot.dev) course curriculum. It follows the [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110) specification for HTTP/1.1.

While the implementation may not be perfect, the goal was to deepen my understanding of how HTTP works under the hood.

## TLS Support

Originally, I planned to implement TLS from scratch. However, I quickly realised the complexity involved in building a secure and standards-compliant TLS implementation. For now, the server uses Go's built-in [`crypto/tls`](https://pkg.go.dev/crypto/tls) package for TLS support.

In the future, I hope to gradually build my own TLS stack as a learning exercise.

