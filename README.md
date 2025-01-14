# kubectl-yadt

Yet Another Diff Tool for Kubernetes - Watch and diff resources in real-time.

> This project is based on [ibuildthecloud/wtfk8s](https://github.com/ibuildthecloud/wtfk8s) with significant improvements.

## Features

- Watch multiple Kubernetes resources simultaneously
- Interactive resource selection with fuzzy search
- Git-style diff output for changes
- Configurable filters for status and metadata changes
- Support for all watchable Kubernetes resources
- Namespace-aware monitoring

## Installation

```bash
# Download and install the binary
go install github.com/futuretea/kubectl-yadt@latest
```

## Usage

### Basic Usage

```bash
# As kubectl plugin
kubectl yadt watch pods

# As standalone command
kubectl-yadt watch pods

# Watch multiple resources
kubectl yadt watch pods deployments services

# Interactive resource selection
kubectl yadt watch
```

### Options

```bash
# Watch resources in a specific namespace
kubectl yadt watch pods --namespace kube-system

# Ignore status changes
kubectl yadt watch pods --no-status

# Ignore metadata changes
kubectl yadt watch pods --no-meta

# Enable debug logging
kubectl yadt watch pods --debug

# Use specific kubeconfig
kubectl yadt watch pods --kubeconfig ~/.kube/other-config

# Use specific context
kubectl yadt watch pods --context other-context
```

### List Available Resources

```bash
# Show all resources that can be watched
kubectl yadt resources
```

## Interactive Mode

When running `kubectl yadt watch` without arguments, you'll enter interactive mode:

1. Use ↑/↓ arrows to navigate resources
2. Type to filter resources
3. Press Enter to select a resource
4. Select multiple resources as needed
5. Choose "Done" to start watching
6. Press Ctrl+C to exit

## Output Format

Changes are displayed in a git-diff style format:

```diff
02:15:30 diff pod.v1 default/nginx-7875f55f56-xk2p4
--------------------------------------------------------------------------------
+ spec:
+   containers:
+     - image: nginx:1.19
+       name: nginx
+       ports:
+         - containerPort: 80
```

## License

Apache License 2.0
