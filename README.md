# Terracotta

Terracotta is an experimental preprocessor for Terraform configurations.

There are several capabilities in Terraform that make it possible to write reusable configurations, including variables and modules.
A preprocessor adds conditional configuration capabilities.

## Example

In this example, an AWS Elastic Load Balancer is being configured.
It's possible that in some cases, SSL might not be used.
We've added a conditional `!if` directive, predicated on `SSL`, to make this part of the configuration optional.

```
resource "aws_elb" "api_elb" {
  name = "api-elb"
  availability_zones = ["us-west-2a", "us-west-2b", "us-west-2c"]
  ...

  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 80
    lb_protocol = "http"
  }

!if SSL
  listener {
    instance_port = 8000
    instance_protocol = "http"
    lb_port = 443
    lb_protocol = "https"
    ssl_certificate_id = "${var.ssl_cert_id}"
  }
!endif
}
```

This simplistic example is somewhat contrived, but it's easy to see how this configuration could be more useful if it were part of a reusable module.

## Syntax

The preprocessor allows symbols to be defined and conditionals to be evaluated against them.
Symbols are either defined or undefined; they may not take on values.
There is no support for macros.

### Directives

The preprocessor recognizes the following directives:

* !define
* !undef
* !if
* !elif
* !else
* !endif

The `!` prefix is used because the `#` character is used for single-line comments in Terraform.

### Expressions

Conditional directives may use Boolean expressions.
These are based on defined symbols.
They include *and* `&&`, *or* `||`, and *grouping* `()` operators.
These may be used with the two conditional directives `!if` and `!elif`.

## Files

Two new file types are used by Terracotta.
Their purpose is to allow Terraform configuration files and the Terraform process to remain unaffected.

### Templates

Terraform uses the `.tf` file extension for configuration files.
Terracotta templates use the `.tft` extension. 

A separate template file is used to simplify interoperability with Terraform.
Since Terraform doesn't recognize preprocessing directives, keeping them in a separate file allows the template to coexist with the corresponding configuration file.

When Terracotta is invoked, it will generate a configuration file corresponding to the associated template file in the same directory.
(This behavior can be changed with the `-output` command-line argument.)
When a `.tft` file is present, the template file can be considered the source, while the associated `.tf` file becomes a generated file.

### Definitions

Preprocessor symbols may be defined in three places.

* Definitions file
* Command line
* Template files

If a `terraform.tfdefs` file is found in a given directory, it will be loaded prior to the templates.
A definitions file is conceptually similar to `terraform.tfvars`, but is only used during preprocessing.

The definitions file is not a template file.
As such, it does not generate a corresponding output file.
Likewise, `.tfvars` files may not contain preprocessing directives.

### JSON

Terracotta only processes Terraform configuration files.
Terraform files are line oritented.
This makes them amenable to preprocessing.
By contrast, JSON files are collection oriented, with comma separators.
In particular, the JSON spec does not allow trailing commas.
While this issue does not rule out preprocessing, it does make preprocessing more problemmatic.

## Command line

Direcives may be defined on the command line.
These will override any directives specified in the definitions file.

```
terracotta -define ECS -define ECS_ELB -undef RDS
```

Command line definitions override those in `terraform.tfdefs`.

## License

Licensed under the Mozilla Public License, like Terraform.
