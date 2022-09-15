# Bill

A command-line application to track job commissions and accounting

## Commands

`bill new client` - add new client to the db (interactive)

`bill new job <CLIENT> <PROJECT>` - start a new job. If `client` is not found, list possible matches.

`bill invoice <JOB_DIR>` - create an invoice from a job and export it to pdf (`cd PROJECT_DIR` and run on `.` by default)

`bill pdf <INVOICE_NUMBER>` - export an invoice to pdf

`bill status [TODO|SENT|...]` - show status for jobs and invoices

`bill all` - print a table of all the registry

`bill filter <FIELD=REGEXP> [FIELD=REGEXP ...]` - filter registry by a certain criteria

`bill get <INVOICE_NUM|PROJECT_NAME|CLIENT>` 

`bill edit <INVOICE_NUM|PROJECT_NAME|CLIENT>` - open registry file in editor and put cursor at the block that matches the given job

