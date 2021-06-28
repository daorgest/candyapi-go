# candyapi-go

A rest api thats made in go with 0 External libararies :D

## Usage

Linux/Unix:

```bash
go run server.go
```

Windows:

```powershell
$Env:ADMIN_PASSWORD = 'anypasswordwillwork';  go run server.go
```

## Envoriment variables

```bash
ADMIN_PASSWORD string
```

## Data types for entry and info

```json
{
    "id": "any id",
    "Name": "Name of the candy",
    "Type": "Type of the candy",
}
```

## Requests you can use

`GET`
`POST`
