# flow-report

Git Flow report tool to show which feature branches have been merged into the develop branch and which ones are stale and can be deleted.

### Installation

Find the executable for your OS in the bin folder and drop it into some place where it's on the path.

### Usage

In your terminal, cd to your development folder and run:

```bash
$ flow-report
```

Output should be something like this:

```bash
Feature "my-feature" exists on 1 repos
	[my-repo] Stale Last activity at 2013-03-18 17:28:07 +0000 GMT
	Feature my-feature has not been merged

Feature "my-other-feature" exists on 8 repos
	[my-repo]
	[config] PR merged at 2014-01-27 12:49:18 +0000 GMT
	[js-repo] PR merged at 2014-02-03 11:48:11 +0000 GMT
	[backend]
	Feature my-other-feature is partially merged
```
