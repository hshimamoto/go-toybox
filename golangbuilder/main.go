// golangbuilder
// MIT License Copyright(c) 2018, 2019, 2020 Hiroshi Shimamoto
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
    "fmt"
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

func loadEnvFromFile(path string) ([]string, error) {
    f, err := os.Open(path)
    if err != nil {
	return nil, err
    }
    defer f.Close()

    envs := []string{}
    lines, err := ioutil.ReadAll(f)
    for _, l := range strings.Split(string(lines), "\n") {
	if strings.Index(l, "=") > 0 {
	    envs = append(envs, l)
	}
    }

    log.Printf("found %s\n", path)

    return envs, nil
}

func loadEnv() []string {
    envs, err := loadEnvFromFile(".golangbuilder.env")
    if err == nil {
	return envs
    }
    home := os.Getenv("HOME")
    envs, err = loadEnvFromFile(filepath.Join(home, ".golangbuilder.env"))
    if err != nil {
	return nil
    }
    return envs
}

func main() {
    cwd, _ := os.Getwd()
    projname := filepath.Base(cwd)
    log.Println("Dir:", cwd, projname)
    gituser := getgituser()
    log.Println("User:", gituser)

    envs := loadEnv()

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
    cname := fmt.Sprintf("gobuild-%s-%d", projname, os.Getpid())

    args := []string{
	"--name", cname, "--hostname", cname,
	"-u", ugid,
	"-v", tmpdir + ":/go",
	"-v", cwd + ":" + dockerwd,
	"-e", "HOME=" + dockerhome,
	"-w", dockerwd,
    }

    for _, env := range envs {
	args = append(args, "-e", env)
    }

    docker(args)
}
