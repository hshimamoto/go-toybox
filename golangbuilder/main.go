// golangbuilder
// MIT License Copyright(c) 2018 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./golangbuilder
//
// bootstrap
// docker run -it --rm \
//   -v $PWD:/go/src/golangbuilder \
//   -u $(id -u):$(id -g) -e HOME=/tmp/go golang bash

package main

import (
    "bufio"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"
)

func readlines(path string) []string {
    lines := []string{}
    f, err := os.Open(path)
    if err != nil {
	return lines
    }
    defer f.Close()
    s := bufio.NewScanner(f)
    for s.Scan() {
	lines = append(lines, s.Text())
    }
    return lines
}

func setup(dir string) {
    // create bin and src
    os.MkdirAll(filepath.Join(dir, "bin"), os.FileMode(0755))
    os.MkdirAll(filepath.Join(dir, "src"), os.FileMode(0755))
}

func docker(args []string) {
    cmd := exec.Command("docker", "run", "-it", "--rm")
    cmd.Args = append(cmd.Args, args...)
    cmd.Args = append(cmd.Args, "golang", "bash")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    err := cmd.Run()
    if err != nil {
	log.Println(err)
    }
}

func getgituser() string {
    name := os.Getenv("USER")
    gitconfig := readlines(".git/config")
    for _, line := range gitconfig {
	// email
	if strings.Index(strings.TrimSpace(line), "email") == 0 {
	    email := strings.Split(line, "=")[1]
	    return strings.TrimSpace(strings.Split(email, "@")[0])
	}
    }
    return name
}

func main() {
    cwd, _ := os.Getwd()
    projname := filepath.Base(cwd)
    log.Println("Dir:", cwd, projname)
    gituser := getgituser()
    log.Println("User:", gituser)

    user, _ := user.Current()
    uid := user.Uid
    gid := user.Gid
    ugid := uid + ":" + gid
    tmpdir, err := ioutil.TempDir("", "golangbuild")
    if err != nil {
	log.Fatal(err)
	os.Exit(1)
    }
    log.Println("TempDir:", tmpdir)
    defer os.RemoveAll(tmpdir)

    setup(tmpdir)

    // default home is /go/src/github.com/<gituser>
    home := filepath.Join("src", "github.com", gituser)
    hosthome := filepath.Join(tmpdir, home)
    dockerhome := filepath.Join("/go", home)
    log.Println("Hosthome:", hosthome)
    os.MkdirAll(hosthome, os.FileMode(0755))
    dockerwd := filepath.Join(dockerhome, projname)
    cname := "gobuild-" + projname

    args := []string{
	"--name", cname, "--hostname", cname,
	"-u", ugid,
	"-v", tmpdir + ":/go",
	"-v", cwd + ":" + dockerwd,
	"-e", "HOME=" + dockerhome,
	"-w", dockerwd,
    }
    docker(args)
}
