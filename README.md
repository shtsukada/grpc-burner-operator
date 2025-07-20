# grpc-burner-operator

Kubernetes上でgRPC負荷生成とObservability統合を行うためのカスタムオペレータとして、以下の技術スタックを活用し作成しています。

- Go + Kubebuilder
- Kubernetes Operator開発
- Observability(Prometheus,Grafana,OpenTelemetry)
- GitOps(ArgoCD)
- CI/CD(GitHub Actions)


## 概要
`grpc-burner-operator`はカスタムリソース(CRD)を通じてKubernetes上でgRPCサーバと負荷生成Podを動的に管理します。将来的にObservabilityの統合やGitOps連携も実装します。


