env "generate" {
  src = "file://schema.hcl"
  dev = "docker://postgres/16/dev?search_path=public"

  format {
    schema {
      inspect = file("models.tmpl")
    }
  }
}
