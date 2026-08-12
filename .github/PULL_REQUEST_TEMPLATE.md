## Summary

<!-- Brief description of what this PR changes and why -->

## Type of Change

- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation
- [ ] Refactor

## Checklist

- [ ] Tests pass in both JSON modes (`go test ./... -race` and `GOEXPERIMENT=jsonv2 go test ./... -race`)
- [ ] Lint passes (`golangci-lint run ./...` and `golangci-lint run --build-tags goexperiment.jsonv2 ./...`)
- [ ] Added/updated tests for new behavior
- [ ] Updated relevant documentation
