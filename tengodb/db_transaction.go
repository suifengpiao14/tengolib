package tengodb

import (
    _ "embed"
)

//go:embed db_transaction.tengo
var TengoDBSource string
