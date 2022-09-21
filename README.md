# Bill

A command-line application to track job commissions and accounting

## Commands

`bill client <CLIENT>` - get client info or add new client to the db if client doesn't exist (interactive)

`bill job <CLIENT> <PROJECT>` - start a new job. If `client` is not found, list possible matches.

`bill jobs` - list open (todo) jobs

`bill invoice <JOB_DIR>` - create an invoice from a job and export it to pdf (`cd PROJECT_DIR` and run on `.` by default)

`bill pdf <INVOICE_NUMBER>` - export an invoice to pdf

`bill status [TODO|SENT|...]` - show status for jobs and invoices

`bill all` - print a table of all the registry

`bill filter <FIELD=REGEXP> [FIELD=REGEXP ...]` - filter registry by a certain criteria

`bill get <INVOICE_NUM|PROJECT_NAME|CLIENT>` 

`bill edit <INVOICE_NUM|PROJECT_NAME|CLIENT>` - open registry file in editor and put cursor at the block that matches the given job

`bill iva <TRIMESTER[1-4]> [YEAR]` - print a table with the accounting results for the given trimester.

