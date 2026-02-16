\ nForth Example: Secure Data Pipeline

\ This example demonstrates fetching data and saving it to a file
\ requires HTTP-ENV and FS-ENV tokens.

ADDRESS HTTP-ENV "http-token"
    "http://google.com" FETCH INTO html
<EXIT>

html "google" CONTAINS INTO is_google

is_google IF
    ADDRESS FS-ENV "fs-token"
        html "google_home.html" WRITE-FILE
    <EXIT>
    "Successfully saved Google home page." PRINT
ELSE
    "Failed to find expected content." PRINT
THEN
