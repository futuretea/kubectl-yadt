# wtfk8s

A command-line tool for watching Kubernetes resources and showing changes in real-time.

## Features

- Watch multiple Kubernetes resources simultaneously
- Interactive resource selection with fuzzy search
- Git-style diff output for changes
- Configurable filters for status and metadata changes
- Support for all watchable Kubernetes resources
- Namespace-aware monitoring

## Installation

```bash
go install github.com/futuretea/wtfk8s@latest
```

## Usage

### Basic Usage

```bash
# Watch a single resource
wtfk8s watch pods

# Watch multiple resources
wtfk8s watch pods deployments services

# Interactive resource selection
wtfk8s watch
```

### Options

```bash
# Watch resources in a specific namespace
wtfk8s watch pods --namespace kube-system

# Ignore status changes
wtfk8s watch pods --no-status

# Ignore metadata changes
wtfk8s watch pods --no-meta

# Enable debug logging
wtfk8s watch pods --debug

# Use specific kubeconfig
wtfk8s watch pods --kubeconfig ~/.kube/other-config

# Use specific context
wtfk8s watch pods --context other-context
```

### List Available Resources

```bash
# Show all resources that can be watched
wtfk8s resources
```

## Interactive Mode

When running `wtfk8s watch` without arguments, you'll enter interactive mode:

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
