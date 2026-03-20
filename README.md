# docs.svc.plus

Go service for Cloud-Neutral documentation delivery and `docs-agent` actions.

## Environment

- `KNOWLEDGE_REPO_PATH`: local checkout of the `knowledge` repository
- `DOCS_SERVICE_PORT`: HTTP listen port
- `INTERNAL_SERVICE_TOKEN`: shared service-to-service auth token
- `DOCS_RELOAD_INTERVAL`: background reload interval, for example `5m`

## Endpoints

- `GET /docs`
- `GET /docs/{collection}`
- `GET /docs/{collection}/{slugPath}`
- `GET /healthz`
- `GET /api/v1/docs/home`
- `GET /api/v1/docs/collections`
- `GET /api/v1/docs/pages/{collection}/{slugPath}`
- `GET /api/v1/blogs`
- `GET /api/v1/blogs/{slugPath}`
- `GET /api/v1/home/latest-blogs`
- `POST /api/v1/admin/reload`
- `POST /api/v1/agent/invoke`

All `/api/v1/*` endpoints require `X-Service-Token`.
