// gradlebuilder
// MIT License Copyright(c) 2020 Hiroshi Shimamoto
// vim:set sw=4 sts=4:
//
// ./gradlebuilder

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
    cmd.Args = append(cmd.Args, "gradle", "bash")
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
    envs, err := loadEnvFromFile(".gradlebuilder.env")
    if err == nil {
	return envs
    }
    home := os.Getenv("HOME")
    envs, err = loadEnvFromFile(filepath.Join(home, ".gradlebuilder.env"))
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
    tmpdir, err := ioutil.TempDir("", "gradlebuild")
    if err != nil {
	log.Fatal(err)
	os.Exit(1)
    }
    log.Println("TempDir:", tmpdir)
    defer os.RemoveAll(tmpdir)

    setup(tmpdir)

    // default home is /home/<gituser>
    home := filepath.Join("/home", gituser)
    log.Println("Hosthome:", tmpdir)
    dockerwd := filepath.Join(home, projname)
    cname := fmt.Sprintf("gradlebuild-%s-%d", projname, os.Getpid())

    args := []string{
	"--name", cname, "--hostname", cname,
	"-u", ugid,
	"-v", tmpdir + ":" + home,
	"-v", cwd + ":" + dockerwd,
	"-e", "HOME=" + home,
	"-w", dockerwd,
    }

    for _, env := range envs {
	args = append(args, "-e", env)
    }

    docker(args)
}
