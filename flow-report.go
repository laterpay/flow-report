package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/wsxiaoys/terminal/color"
	"io/ioutil"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	gitPath   string
	repos     = []Repo{}
	re        = regexp.MustCompile(`\s[a-f0-9]{7}\s`)
	staleDate = time.Now().Truncate(time.Duration(time.Hour * 24 * 30 * 2))
	verbose   = false
)

type featureBranch struct {
	Name           string
	FullName       string
	LastCommit     string
	LastCommitDate time.Time
}

type branchMerge struct {
	CommitHash  string
	AuthorName  string
	AuthorEmail string
	Message     string
	Branch      *featureBranch
	Date        time.Time
}

type Repo struct {
	Path            string
	MergeLog        []branchMerge
	featureBranches []featureBranch
}

func (r *Repo) getBranch(name string) *featureBranch {
	for idx, b := range r.featureBranches {
		if b.Name == name {
			return &r.featureBranches[idx]
		}
	}
	return nil
}

func NewRepo(path string) Repo {
	return Repo{
		Path:     path,
		MergeLog: []branchMerge{},
	}
}

type Context struct {
	Repos                []Repo
	featureBranchOnRepos map[string][]*Repo
}

func init() {
	var err error
	gitPath, err = exec.LookPath("git")
	if err != nil {
		log.Fatal("Git not found")
	}
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()
}

func parseBranches(output string) (ri []featureBranch) {

	for _, s := range strings.Split(output, "\n") {

		s = strings.Trim(s, " ")
		if s == "" || !strings.Contains(s, "feature") {
			continue
		}

		splitPos := re.FindStringIndex(s)
		lastCommit := strings.Trim(s[splitPos[0]:splitPos[1]], " ")

		s = strings.Trim(s[:splitPos[0]], " ")
		fullName := strings.TrimLeft(s, "* ")
		n := strings.Split(fullName, "/")

		fb := featureBranch{
			Name:       n[len(n)-1],
			FullName:   fullName,
			LastCommit: lastCommit,
		}

		ri = append(ri, fb)
	}

	return ri

}

func parseLog(output string, repo *Repo) (bm []branchMerge) {

	for _, l := range strings.Split(output, "\n") {

		if l == "" {
			continue
		}

		vars := strings.Split(l, ";")
		msg := vars[3]
		if strings.HasPrefix(msg, "Merge branch 'develop'") {
			continue
		}

		t, _ := strconv.ParseInt(strings.Trim(vars[4], " '"), 10, 64)

		y := branchMerge{
			CommitHash:  vars[0],
			AuthorName:  vars[1],
			AuthorEmail: vars[2],
			Message:     msg,
			Date:        time.Unix(t, 0),
		}

		for idx, fb := range repo.featureBranches {
			if strings.Contains(y.Message, fb.Name) {
				y.Branch = &repo.featureBranches[idx]
			}
		}

		bm = append(bm, y)

	}

	return bm
}

func runCommand(path, command, args string) (string, error) {
	if verbose {
		log.Printf("%s %s %s\n", path, command, args)
	}
	cmd := exec.Command(command, strings.Split(args, " ")...)
	cmd.Dir = path
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func inspectRepository(repo *Repo) {

	output, err := runCommand(repo.Path, gitPath, "fetch")
	if err != nil {
		log.Fatal(err)
	}
	output, err = runCommand(repo.Path, gitPath, "branch -a -v")
	if err != nil {
		log.Fatal(err)
	}
	repo.featureBranches = parseBranches(output)

	//todo on demand
	var wg sync.WaitGroup
	for idx := range repo.featureBranches {
		wg.Add(1)
		go func(i *featureBranch) {
			defer wg.Done()
			output, err = runCommand(repo.Path, gitPath, fmt.Sprintf("show %s --pretty=format:'%%ct'", i.LastCommit))
			t, _ := strconv.ParseInt(strings.Trim(strings.Split(output, "\n")[0], " '"), 10, 64)
			i.LastCommitDate = time.Unix(t, 0)
		}(&repo.featureBranches[idx])
	}
	wg.Wait()

	output, err = runCommand(repo.Path, gitPath, "log --merges --grep=feature --pretty=format:'%H;%an;%ae;%s;%ct'")
	if err != nil {
		log.Fatal(err)
	}

	repo.MergeLog = parseLog(output, repo)

}

func buildContext() Context {

	context := Context{
		Repos:                []Repo{},
		featureBranchOnRepos: make(map[string][]*Repo),
	}

	temps, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal("Repohome invalid")
	}
	for _, f := range temps {
		if f.IsDir() {
			context.Repos = append(context.Repos, NewRepo(f.Name()))
		}
	}

	var wg sync.WaitGroup
	for idx := range context.Repos {
		wg.Add(1)
		go func(repo *Repo) {
			defer wg.Done()
			inspectRepository(repo)
		}(&context.Repos[idx])
	}
	wg.Wait()

	for idx, r := range context.Repos {
		for _, f := range r.featureBranches {
			_, ok := context.featureBranchOnRepos[f.Name]
			if !ok {
				context.featureBranchOnRepos[f.Name] = []*Repo{}
			}
			context.featureBranchOnRepos[f.Name] = append(context.featureBranchOnRepos[f.Name], &context.Repos[idx])
		}
	}

	return context
}

func main() {

	context := buildContext()

	for f, r := range context.featureBranchOnRepos {

		color.Print("@g", fmt.Sprintf("Feature \"%s\" exists on %d repos\n", f, len(r)))
		partiallyMerged := false
		anyMerge := false

		for _, r := range r {
			color.Print("@y", fmt.Sprintf("\t[%s]", r.Path))
			repoMerged := false
			for _, m := range r.MergeLog {
				if m.Branch != nil && m.Branch.Name == f {
					anyMerge = true
					repoMerged = true
					fmt.Printf(" PR merged at %s", m.Date)
				}
			}
			if !repoMerged {

				branchOnRepo := r.getBranch(f)
				if branchOnRepo != nil {
					if branchOnRepo.LastCommitDate.Before(staleDate) {
						color.Print("@r", " Stale ")
						color.Print("@c", fmt.Sprintf("Last activity at %s", branchOnRepo.LastCommitDate))
					}
				}
				partiallyMerged = true

			}

			fmt.Printf("\n")
		}

		if partiallyMerged && anyMerge {
			color.Print("@r", fmt.Sprintf("\tFeature %s is partially merged\n", f))
		}
		// no merges
		if !anyMerge {
			color.Print("@b", fmt.Sprintf("\tFeature %s has not been merged\n", f))
		}

		if !partiallyMerged {
			color.Print("@c", fmt.Sprintf("\tFeature %s is fully merged\n", f))
		}

		fmt.Printf("\n")

	}

}
