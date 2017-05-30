# KUBE-WATCH
Kube-watch is a Kubernetes BOT to spy for Crashed Pods.

## How to run it
### Create a Secret
Create a Kubernetes Secret to hold your Slack token.
```
$ kubectl create secret generic kube-watch --from-literal=token=<slack-token>
```

### Create Kube-Watch deployment.yaml
```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: kube-watch
  labels:
    component: kube-watch
spec:
  replicas: 1
  template:
    metadata:
      labels:
        component: kube-watch
    spec:
      containers:
      - name: kube-watch
        image: drgarcia1986/kube-watch:0.0.3
        imagePullPolicy: Always
        env:
        - name: K8SENV
          value: <your-k8s-environment-label>
        - name: CIRCLETIME
          value: "5"
        - name: SLACKAVATAR
          value: <slack-avatar-for-bot>
        - name: SLACKCHANNEL
          value: <slack-channel>
        - name: SLACKTOKEN
          valueFrom:
            secretKeyRef:
              name: kube-watch
              key: token
```
### Make deploy
```
$ kubectl create -f deployment.yaml
```
