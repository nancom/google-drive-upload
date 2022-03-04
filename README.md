# Golang Google Drive Upload

## Prepare Credentail and Token

### Prepare .env

```env
UPLOAD={upload folder id}
SRC={upload src file}
FILE_LIST={file list , separate}
```

## Run

```bash
$export $(cat .env | xargs)
$go run . {sub-folder}
```

```bash
$ env GOOS=linux GOARCH=amd64 go build
```

## refrence

* [Google Drive API v3](https://developers.google.com/drive/api/v3/quickstart/go)
* [Google create credentail : desktop](https://developers.google.com/workspace/guides/create-credentials#desktop-app)
