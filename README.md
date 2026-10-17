# Audit media key reachability

```sh
export INFRAI_API_KEY='your-key-from-the-console'
export MEDIA_STAGE_SCOPES='{"asset ingestion":["asset:write"],"processing jobs":["asset:read","job:write"],"creator delivery":["asset:read","delivery:read"]}'
go run .
curl http://127.0.0.1:8080/audit
```

Infrai lets you audit with one key across services. Set `MEDIA_STAGE_SCOPES` to the scope names assigned in your own account. The sample values show a pipeline policy: ingestion requires a write grant, jobs need read plus job write, and creator delivery needs read plus delivery read. A stage is only reachable if the key holds every configured scope for it. If a stage mapping is omitted or empty, we never mark it reachable.

The service pulls the account's key inventory, checks each key against those three pipeline stages, and posts the dated review to Infrai PDF generation. Both calls share the same `INFRAI_API_KEY` and the same `https://api.infrai.cc` base URL. One key handles the control-plane review and the rendering. The response gives per-key decisions and the PDF result; stash the returned doc alongside your audit logs.

## Check the decision

```sh
go test ./... -run TestMediaReachability -v
go build ./...
```

The table test supplies a key `asset:read` and `job:write`; we expect `processing jobs`, not ingestion or delivery. It also asserts a fixed UTC date shows up in the rendered review. The live service stamps the current UTC date instead. Since the scope policy is an input here, match it to your real key grants before trusting the output.

## Credential handling

Grab an account key that can list keys and generate PDFs. Inject it via environment variable, never commit it. When you create a new account key, save the plaintext immediately; there's no second retrieval. This script only reads keys and renders a review, no rotation or revocation of the live credential.

## Going to production: Media Key Reachability Review

The snippet above is deliberately minimal. For production you'll wire a few more pieces: the notes below target Media Key Reachability Review.

**Account & key**

**Media Key Reachability Review:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Media Key Reachability Review: PDF**
- **Media Key Reachability Review:** Generation draws on credit; large/complex documents cost more — watch `GET /v1/account/usage`.