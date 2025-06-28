# values-blame

## Purpose

`values-blame` is a tool designed to help you identify which `values.yaml` file edits a specific line when you have multiple `values.yaml` files in your Helm install command. This is particularly useful for debugging and understanding the impact of multiple configuration files in complex Helm deployments.

## Build

```
go build
```

## Usage

To use `values-blame`, simply provide the `values.yaml` files as input using the `-f` flag. For example:

```bash
./values-blame -f values.yaml -f values-2.yaml
```

Output
```
values-2.yaml global: 
values-2.yaml   oidc: 
values-2.yaml     enabled: true
values.yaml     tls: 
values.yaml       enabled: true
```