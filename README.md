
# AutoREST

This package is a simple system for automatically making REST endpoints from an arbitrary structure stored in a GORM
database. The core package gives you a simple adapter type that handles all the mechanics of working with the database.
The sub packages all use this API to make ready to use HTTP handlers for a few different HTTP router packages.

Obviously this system is based on heavy use of refection, and it isn't the most flexible design ever. A dedicated
handmade API will always be better. However... When you just need some endpoints you can hit and get data, this is fast
and easy.
