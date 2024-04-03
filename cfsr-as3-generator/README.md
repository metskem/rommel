### cfsr-as3-generator

Cloud Foundry Smart Routing AS3 generator.

A sample piece of code that generates [AS3](https://clouddocs.f5.com/products/extensions/f5-appsvcs-extension/latest/) json for a given number of apps.  
AS3 is a declarative interface to configure BIG-IP devices.  

The idea is that we generate that config based on pieces of template for CF apps, so that each app has it's own F5 pool, healthcheck and entry in a datagroup.


