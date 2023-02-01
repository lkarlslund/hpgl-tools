function BuildVariants {
  param (
    $compileflags,
    $prefix,
    $suffix,
    $arch,
    $os,
    $path
  )

  foreach ($currentarch in $arch) {
    foreach ($currentos in $os) {
      $env:GOARCH = $currentarch
      $env:GOOS = $currentos
      go build -o binaries/$prefix-$currentos-$currentarch-$VERSION$suffix $compileflags $path
    }
  }
}

Set-Location $PSScriptRoot

$COMMIT = git rev-parse --short HEAD
$VERSION = git describe --tags --exclude latest
$DIRTYFILES = git status --porcelain

if ("$DIRTYFILES" -ne "") {
  $VERSION = "$VERSION-local-changes"
}

# Release
BuildVariants -prefix hpgl-optimizer -path ./optimizer -arch @("amd64", "arm64") -os @("windows") -suffix ".exe"
BuildVariants -prefix hpgl-optimizer -path ./optimizer -arch @("amd64", "arm64") -os @("darwin", "linux")

BuildVariants -prefix hpgl-sender -path ./sender -arch @("amd64", "arm64") -os @("windows") -suffix ".exe"
BuildVariants -prefix hpgl-sender -path ./sender -arch @("amd64", "arm64") -os @("darwin", "linux")
