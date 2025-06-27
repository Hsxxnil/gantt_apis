# PMIP APIs

一套以 **Golang** + **PostgreSQL** 為後端、**Angular** 為前端框架開發的 **甘特圖式專案管理平台**。系統支援使用者帳號註冊與登入，並提供專案、任務及資源等的完整 CRUD 功能。  透過視覺化甘特圖介面，協助用戶清楚規劃專案進度、有效分配人力與資源。

## 專案連結

* 前端畫面：[點我查看](http://hsxxnil.notion.site/PMIP-11c5b51f95f5816e904ec80bdb5a9023)
* Swagger API 文件：[點我查看](https://hsxxnil.github.io/swagger-ui/)

## 安裝
1. 下載專案

```bash
git clone https://github.com/Hsxxnil/gantt_apis.git
cd gantt_apis
```

2. 建立 Makefile

> 請根據您的作業系統選擇對應的範本進行複製：
* Windows
```bash
copy Makefile.example.windows Makefile
```
* Linux / macOS
```bash
cp Makefile.example.linux Makefile
```

3. 初始化

> 執行以下指令將自動安裝依賴套件並建立必要的目錄結構：
```bash
make setup
```
4. 設定環境參數
> 開啟並編輯以下檔案，填入資料庫連線資訊、JWT 金鑰等必要參數：
```file
config/debug_config.go
```

## 資料庫遷移

> 執行以下指令使用[golang-migrate](https://github.com/golang-migrate/migrate)做資料庫遷移及做資料表版控：
```bash
make migration
```

## 執行
> 執行以下指令在本地端啟動伺服器並自動重載：
```bash
make air
```

## License

本專案使用的 [Vodka](https://github.com/dylanlyu/vodka) 採用 [MIT License](https://opensource.org/licenses/MIT) 授權。
