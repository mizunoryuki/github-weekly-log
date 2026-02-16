# GitHub weekly log

このリポジトリは個人のGitHubアカウントの活動履歴を調べ、特定の日時(毎週土曜日朝9時)に活動履歴をまとめたメールをユーザに届ける

## リポジトリ内のディレクトリ構造

このリポジトリは3つのプロジェクトが配置されている

- Go
- Hono
- Astro

### Go(メインロジック)

このプロジェクトでは主に以下のことを行う

- GitHub APIからUserのコミット情報を取得
  - 先週と今週のデータを保持する
- コミット情報のJSONファイルを生成
- Cloudflare D1 にデータ保存
- ユーザにメール送信

### Hono(worker API)

Astroのプロジェクトで表示するためのデータをD1からフェッチするためのAPI
一週間ごとのデータを履歴として閲覧することができる

### Astro(pages フロント)

HonoのAPIを通じて過去のコミット情報を取得して一覧表示するためのフロントエンド部分

現在実装できている部分

- 一週間のデータをグラフを表示

今後実装したい部分

- 期間比較
- 全期間比較

## 技術構成

### Go

go-github
cloudflare-go
resend-go
mjml

Resendというメール配信サービスを利用し、指定したメールアドレスにメールを送信する
送信する内容はmjmlによって作成してい

mjmlはHTMLのメールを簡単に記述するためのフレームワークで、メールの配信先はGmailとしていたのでHTMLによるメールコンテンツの作成を行なった

<img width="348" alt="mail template" src="https://github.com/user-attachments/assets/a7f27772-b7cf-4cd4-a564-133e3c6c5814" />

cloudflare D1とJSONによるデータの保存をしている理由として、cloudflare D1はウェブアプリケーションのコンテンツ表示のためのデータベースとしていて、JSONはMac OSのウィジェットにコンテンツを表示するために利用している。
今回のリポジトリではMac OSのウィジェットのプログラムは含まれていないが、cloudflare D1のアクセスが増えることを危惧してアクセスの分散を目的としてJSONによるデータ保存を行っている
作成したJSONファイルは別のリポジトリに保存するようにしている

<img width="348" height="170" alt="スクリーンショット 2026-02-17 0 15 16" src="https://github.com/user-attachments/assets/6a55b35f-149a-4605-bcdf-8f1c814c6981" />


このプログラムはGitHub ActionsのWorkflowによって定期実行されるようになっている
毎週土曜日9時に実行され、一週間分のデータ(実行される前の週の土曜日から実行される週の金曜日までが対象)を取得し、データを保存し、メール配信を行うことを一つのフローとしている。

### Hono

cloudflareworkers-types
wrangler

cloudflare workersによりデプロイを行っている
cloudflare accessを用いてAPIのアクセス制限を設けており、自分以外からのアクセスを制限している

### Astro

wrangler
chartjs

cloudflare pagesによりデプロイを行っている
こちらもHonoと同様にcloudflare accessによるアクセス制限を用いている
このサイトではグラフの表示を行っている
<img width="1043" height="833" alt="スクリーンショット 2026-02-17 0 15 51" src="https://github.com/user-attachments/assets/dc5398d1-9c72-421d-b07d-f98541e5317c" />
