# grpc-burner-operator

## プロジェクト概要
`grpc-burner-operator`はgRPCワークロードの起動、負荷生成、可観測性(Metrics/Logs/Traces)、アラート通知をGitOps(Argo CD)で管理するためのKubernetes Operatorです。

## 目的
- Operator/CRD、GitOps、Observability、Slack 通知、Sealed Secrets を エンドツーエンドで実装例として示す

## 主な機能
- CRD管理
    - GrpcBurner：gRPC サービスの起動・基本設定
    - BurnerJob：負荷生成用 Job の起動・制御
    - ObservabilityConfig：メトリクス/ログ/トレースやアラート閾値の有効化
- GitOps：deploy/argocd の AppProject / App-of-Apps / Applications によりクラスター全体を同期
- 可観測性：config/monitoring による kube-prometheus-stack / Loki / Tempo / Grafana の構築、config/prometheus の ServiceMonitor 連携
- 通知：Argo CD・Grafana・Alertmanager の Slack 通知


カスタムリソース(CRD)を通じてKubernetes上でgRPCサーバと負荷生成Podを動的に管理します。将来的にObservabilityの統合やGitOps連携も実装します。


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

## クイックスタート(EKS/Argo CD/Sealed Secrets)

1) EKS を用意
```bash
cd infra/terraform/eks
terraform init
terraform apply -var-file=terraform.tfvars

# kubeconfig 連携
aws eks update-kubeconfig --name grpc-observability-cluster --region ap-northeast-1
```
2) GitOps 基盤（Argo CD）を導入
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 初期パスワード（UI ログイン時に使用）
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d; echo
```
3) Sealed Secrets を導入
```bash
kubectl apply -f https://github.com/bitnami-labs/sealed-secrets/releases/latest/download/controller.yaml
```

4) AppProject と App-of-Apps を適用（GitOps 起点）
```bash
kubectl apply -f deploy/argocd/project.yaml
kubectl apply -f deploy/argocd/app-of-apps.yaml
```
5) Slack Webhook を SealedSecret 化（通知の事前設定）
```bash
# Argo CD Notifications（argocd ns）
kubectl -n argocd create secret generic argocd-notifications-secret \
  --from-literal=slack-webhook="https://hooks.slack.com/services/XXX/YYY/ZZZ" \
  --dry-run=client -o yaml \
| kubeseal --controller-name=sealed-secrets --controller-namespace kube-system -o yaml \
> deploy/argocd/system/argocd-notifications-secret-sealed.yaml
kubectl apply -f deploy/argocd/system/argocd-notifications-secret-sealed.yaml

# Grafana / Alertmanager（monitoring ns）
kubectl -n monitoring create secret generic grafana-slack-webhook \
  --from-literal=url="https://hooks.slack.com/services/XXX/YYY/ZZZ" \
  --dry-run=client -o yaml \
| kubeseal -n monitoring --controller-name=sealed-secrets --controller-namespace kube-system -o yaml \
> config/monitoring/grafana/grafana-slack-webhook-sealed.yaml
kubectl apply -f config/monitoring/grafana/grafana-slack-webhook-sealed.yaml

kubectl -n monitoring create secret generic alertmanager-slack-webhook \
  --from-literal=url="https://hooks.slack.com/services/XXX/YYY/ZZZ" \
  --dry-run=client -o yaml \
| kubeseal -n monitoring --controller-name=sealed-secrets --controller-namespace kube-system -o yaml \
> config/monitoring/kube-prometheus-stack/alertmanager-slack-webhook-sealed.yaml
kubectl apply -f config/monitoring/kube-prometheus-stack/alertmanager-slack-webhook-sealed.yaml
```

6) サンプル CR を適用（gRPC サービス/負荷/監視）
```bash
kubectl apply -f config/samples/
```
7) 動作確認（UI と通知）
```bash
# Argo CD UI
kubectl -n argocd port-forward svc/argocd-server 8080:80
# -> https://localhost:8080 （admin / 初期PW）

# Grafana（kube-prometheus-stack の初期値に依存）
kubectl -n monitoring port-forward svc/kube-prometheus-stack-grafana 3000:80
# -> http://localhost:3000

# Pod/ターゲットの健全性
kubectl get pods -A
kubectl -n monitoring get servicemonitors,prometheusrules

# （任意）Sync 成功/失敗やテスト用アラートで Slack 通知を確認
```
