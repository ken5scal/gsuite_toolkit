# gsuite_toolkit

# Future Improvement
* Make a New Account
* Register Google Group(Mailing List and Authority Group)
* Link Yubikey with Account
* List Google Forms and check sharing settings(Periodically)
  * If some one creates new doc, then notify to the slack or related chat


# Future ToDO
- ~~カレンダーでリソース管理~~
- グループ管理
 - [x] list
 - [x] get
 - [ ] create
 - [ ] delete
 - [ ] edit
- ユーザー管理
- ドライブの監査
- Adminログイン時の通知
- 特権ユーザー
 - [x] list
 - [x] 権限付与
 - [ ] 権限剥奪
 - [ ] ログイン履歴取得
 - [ ] 通知
- ユーザー
 - [x] list
 - [x] 凍結されてるやつ
 - [ ] 作成
 - [ ] 通知  
 - [x] 2sv
 - [x] リカバリコード
- ユーザーバルク作成
 - ~~[x] 作成~~
 - [ ] Do it using go func. multipart/mix is hard because it is hard to check inner responses status
- 不正ログイン
 - [x] isSuspicious
- ファイル権限
- ファイル共有設定
- [ ] 出力のJson化
- [ ] エラーの出力(JSON)

# Command
* MY_EXECUTABLE [optional flags] <group | commands>
* MY_EXECUTABLE group [optional flags] <group | commands>
* MY_EXECUTABLE group sub_group [optional flags] <command>

