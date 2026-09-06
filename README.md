# go-simply-sdk

A context-aware Go client for the Simply.com API v2.7.0 contract. The client keeps authentication, response metadata, and provider-shaped DNS/product operations in this independently versioned module.

Product listing is the supported account-validation operation. Simply does not expose a verified account identity endpoint in the current contract, so the SDK and future application adapter must not infer identity from products.

```go
client, err := simply.NewClient(simply.Config{APIKey: key})
products, response, err := client.Products().List(ctx)
records, response, err := client.DNS().ListRecords(ctx, productObject)
```

Bearer and Basic authentication, injected HTTP clients, context cancellation, bounded error bodies, and rate-limit metadata are supported. `make check` is offline and requires no credentials. See `testdata/support-matrix.json` for the complete tracked operation set.
