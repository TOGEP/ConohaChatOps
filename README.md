## ConohaChatOps
ConohaChatOps is an application made to manage conoha vps from discord.  
**ConohaChatOps**は[Conoha VPS](ttps://www.conoha.jp/)をDiscord上から管理するために作られたアプリケーションです。

## Requirement
Go1.16 or later

## 環境変数の取得方法

<details open>
<summary>1. BOTTOKEN</summary>
  
[Discord Developer Portalを参考に](https://discord.com/developers/docs/getting-started)アプリケーションを作成後、左メニューの 「Bot」 > TOKENをコピーで取得
</details>

<details open>
<summary>2. GUILDID</summary>
  
特定のサーバのみで稼働させたい場合にのみ指定するオプション項目です。  
[discordのサポートページを参考に](https://support.discord.com/hc/ja/articles/206346498)デベロッパモードを有効化後、Botを稼働させたいサーバーを右クリック >「IDをコピー」 で取得
</details>

<details open>
<summary>3. Conoha_ENDPOINT</summary>
  
左メニューの「API」を選択

![Get-ConohaEndpoint1](docs/img/Get-ConohaEndpoint1.png)

「エンドポイント」をクリックして表示されるURLからエンドポイントのリージョンを取得

![Get-ConohaEndpoint2](docs/img/Get-ConohaEndpoint2.png)
</details>

<details open>
<summary>4. Conoha_TENANTID</summary>
  
「API」画面内の「テナント情報」をクリックしてテナントIDを取得

![Get-ConohaTenantID](docs/img/Get-ConohaTenantID.png)
</details>

<details open>
<summary>5. Conoha_TENANTNAME</summary>
  
「API」画面内の「テナント情報」をクリックしてテナント名を取得

![Get-ConohaTenantName](docs/img/Get-ConohaTenantName.png)
</details>

<details open>
<summary>6. Conoha_USERNAME</summary>
  
[Conohaの利用ガイドを参考に](https://support.conoha.jp/v/addapiuser/)APIユーザを追加後、ユーザ名を取得

![Get-ConohaAPIUserName](docs/img/Get-ConohaAPIUserName.png)
</details>

<details open>
<summary>7. Conoha_PASSWORD</summary>
  
6の作業で作成したAPIユーザのパスワードを取得
</details>

## ローカルでBotを稼働する場合

## HerokuでBotを稼働する場合
