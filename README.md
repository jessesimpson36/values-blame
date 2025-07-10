# values-blame

## Purpose

`values-blame` is a Helm plugin designed to help you identify which `values.yaml` file edits a specific line when you have multiple `values.yaml` files in your Helm install command. This is particularly useful for debugging and understanding the impact of multiple configuration files in complex Helm deployments.

## Installation

Install the plugin directly from the repository:

```bash
helm plugin install https://github.com/your-username/values-blame
```

Or install locally if you've cloned the repository:

```bash
helm plugin install .
```

## Build (for development)

```bash
go build
```

## Usage

Use the plugin with Helm to analyze multiple values files:

```bash
helm values-blame -f values.yaml -f values-2.yaml
```

### Options

- `-f` or `--values`: Specify values files (can be used multiple times)
- `-c`: Only print coalesced values
- `-n`: Don't print file names in output

### Examples

Basic usage with two values files:
```bash
helm values-blame -f values.yaml -f values-2.yaml
```

Show only the final coalesced values:
```bash
helm values-blame -f values.yaml -f values-2.yaml -c
```

Hide file names in output:
```bash
helm values-blame -f values.yaml -f values-2.yaml -n
```

## Output

The plugin shows which file each configuration value comes from:

```
values-2.yaml global: 
values-2.yaml   oidc: 
values-2.yaml     enabled: true
values.yaml     tls: 
values.yaml       enabled: true
```

This makes it easy to trace configuration overrides and understand the final merged configuration.