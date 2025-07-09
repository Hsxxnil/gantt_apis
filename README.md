# PMIP APIs

一套以 **Golang** + **PostgreSQL** 為後端、**Angular** 為前端框架開發的 **甘特圖式專案管理平台**，平台核心功能為將專案中各項任務流程視覺化，透過動態甘特圖呈現，有效掌握專案全貌與進度。
此系統可協助使用者：
* 即時追蹤任務進展與期限
* 明確分配與管理人力、資源
* 預測潛在風險並進行調整
* 提升跨部門協作效率與透明度
  
透過前後端分離架構設計，系統具備良好的擴充性與維護性，致力於打造一個高效、直覺的專案管理工具。

## 專案連結

* 前端畫面：[點我查看](http://hsxxnil.notion.site/PMIP-11c5b51f95f5816e904ec80bdb5a9023)
* Swagger API 文件：[點我查看](https://hsxxnil.github.io/swagger-ui/?urls.primaryName=Gantt)

## 安裝
1. 下載專案

```bash
git clone https://github.com/Hsxxnil/gantt_apis.git
cd gantt_apis
```

2. 建立 Makefile

> 請根據您的作業系統選擇對應的範本進行複製：
* Linux / macOS
```bash
cp Makefile.example.linux Makefile
```

* Windows
```bash
copy Makefile.example.windows Makefile
```

3. 初始化

> 如為初次建立開發環境，請先根據您的作業系統安裝必要套件：
* Linux / macOS
```bash
brew install golang-migrate golangci-lint protobuf
```

* Windows（建議使用 Scoop，或手動安裝以下套件）：
```bash
scoop install golang-migrate golangci-lint protobuf
```

> 執行以下指令將自動安裝依賴套件並建立必要的目錄結構：
```bash
make setup
```

4. 設定環境參數

> 開啟並編輯以下檔案，填入資料庫連線資訊、JWT 金鑰等必要參數：
```file
config/debug_config.go
```

5. 更新套件

>執行以下指令升級相關套件
```bash
make update_lib
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
