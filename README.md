# grpc-burner-operator

Kubernetes上でgRPC負荷生成とObservability統合を行うためのカスタムオペレータとして、以下の技術スタックを活用し作成しています。

- Go + Kubebuilder
- Kubernetes Operator開発
- Observability(Prometheus,Grafana,OpenTelemetry)
- GitOps(ArgoCD)
- CI/CD(GitHub Actions)


## 概要
`grpc-burner-operator`はカスタムリソース(CRD)を通じてKubernetes上でgRPCサーバと負荷生成Podを動的に管理します。将来的にObservabilityの統合やGitOps連携も実装します。



## 運用メモ

- 本構成はArgo CDにより、`App-of-Apps`パターンでgRPCアプリ/監視スタック/Otel Collectorを順序適用します。
- 適用順は`sync-wave`により制御：
  - Wave -2 : monitoring-stack(Prometheus/Loki/Tempo)
  - Wave -1 : otel-collector(外部Helmチャート、自リポジトリvalues)
  - Wave  0 : grpc-burner(Operator、gRPCアプリ)
- 適用手順：
  1. Argo CD が稼働しているクラスタに `project.yaml` を適用
  2. `app-of-apps.yaml` を適用（子Applicationが自動登録・同期）
  ```bash
  kubectl apply -n argocd -f deploy/argocd/project.yaml
  kubectl apply -n argocd -f deploy/argocd/app-of-apps.yaml
  ```