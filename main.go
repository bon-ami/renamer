/* Move a file to a directory, renaming by regular expression.
 If destination not ready, wait.
 The most current file matching the regular expression will be chosen
   and the rest will be removed.
parameters:	--dstDir="destination directory"
		--srcDir="source directory"
		--srcStr="source file name (Regular Expression)"
		--dstStr="part of destination file name"
		--dstAffix="destination file name affix"
		-ignore //not to wait for destination ready
		-debug //more info print and no removal
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var dstDir, srcDir, srcStr, dstStr, dstAffix string
var ignore, dbg bool

// init variables
func initVar() {
	flag.StringVar(&srcDir, "srcDir", "", "source directory")
	flag.StringVar(&dstDir, "dstDir", "", "destination directory")
	flag.StringVar(&srcStr, "srcStr", "", "source filename string")
	flag.StringVar(&dstStr, "dstStr", "", "part of destination filename")
	flag.StringVar(&dstAffix, "dstAffix", "", "destination filename affix")
	flag.BoolVar(&ignore, "ignore", false, "to ignore destination not ready error (default false)")
	flag.BoolVar(&dbg, "debug", false, "no file removal and to print debugging messages (default false)")
	flag.Parse()
}

func wait4dst(dstDir string, ignore bool) (dir []os.FileInfo, ok bool) {
	ok = false
	var err error
	for {
		if dir, err = ioutil.ReadDir(dstDir); err != nil {
			if ignore {
				break
			} else {
				fmt.Println("Make " + dstDir + " ready and press any key.")
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
			}
		} else {
			ok = true
			return
		}
	}
	return
}

func get1src(srcDir string, srcFiles []os.FileInfo) (srcFile string) {
	var tm time.Time
	for _, newFile := range srcFiles {
		if dbg {
			fmt.Printf("Comparing %s to %s\n", newFile.ModTime(), tm)
		}
		if dbg {
			fmt.Printf("%s size is %d\n", newFile.Name(), newFile.Size())
		}
		if newFile.Size() > 0 && newFile.ModTime().After(tm) {
			tm = newFile.ModTime()
			srcFile = srcDir + newFile.Name()
			if dbg {
				fmt.Println(srcFile + " is more recent.")
			}
		}
	}
	return
}

func cp(srcFile, dstFile string) {
	if dbg {
		fmt.Println("to copy " + srcFile + " to " + dstFile)
	}
	if src, err := os.Open(srcFile); err != nil {
		fmt.Println("FAILED to open source file!" + err.Error())
	} else {
		defer src.Close()
		if dst, err := os.Create(dstFile); err == nil {
			defer dst.Close()
			io.Copy(dst, src)
		} else {
			fmt.Println("FAILED to create target file!" + err.Error())
		}
	}
}

func nameDst(dstDir, dstStr, dstAffix string) (dstFile string) {
	return dstDir + dstStr + dstAffix
}

func getsrcs(srcDir, srcStr string) (srcFiles []os.FileInfo) {
	files := []os.FileInfo{}
	var err error
	if files, err = ioutil.ReadDir(srcDir); err != nil {
		fmt.Println("NO source files found!" + err.Error())
		return nil
	}
	srcFiles = []os.FileInfo{}
	for _, file1 := range files {
		if dbg {
			fmt.Println("Comparing " + file1.Name() + " to " + srcStr)
		}
		if matched, err := regexp.MatchString(srcStr, file1.Name()); matched && err == nil {
			srcFiles = append(srcFiles, file1)
			if dbg {
				fmt.Println("Source file " + file1.Name() + " detected")
			}
		}
	}
	if dbg {
		fmt.Print("All source files are ")
		for _, file1 := range srcFiles {
			fmt.Print(file1.Name() + " ")
		}
		fmt.Print(".\n")
	}
	return
}

func rm(dir, file string) {
	fmt.Println("To delete " + dir + file)
	if !dbg {
		os.Remove(dir + file)
	}
}

func dirCorrect(dir string) (string, bool) {
	sep, err := strconv.Unquote(strconv.QuoteRuneToGraphic(os.PathSeparator))
	if err != nil {
		fmt.Println("Failed to check path separator!")
		return "", false
	}
	if strings.HasSuffix(dir, sep) {
		return dir, false
	}
	return dir + sep, true
}

/*
 */
func main() {
	initVar()
	if len(srcDir) < 1 || len(srcStr) < 1 || len(dstStr) < 1 || len(dstDir) < 1 || len(dstAffix) < 1 {
		if dbg {
			fmt.Println("source dir=" + srcDir +
				", source files=" + srcStr +
				", destination dir=" + dstDir +
				", destination files=" + dstStr +
				", destination affix=" + dstAffix)
		}
		fmt.Println("All source and destination, dirs and name patterns needed! Use -h for help.")
		os.Exit(-1)
	}

	srcDir, _ = dirCorrect(srcDir)
	dstDir, _ = dirCorrect(dstDir)
	if dbg {
		fmt.Println("After separator check, source dir=" +
			srcDir + ", destination dir=" + dstDir)
	}
	// source check
	srcFiles := getsrcs(srcDir, srcStr)
	if srcFiles != nil {
		// target check
		dir, ok := wait4dst(dstDir, ignore)
		if ok && dir != nil {
			// check old files
			for _, oldFile := range dir {
				if strings.Contains(oldFile.Name(), dstStr) {
					rm(dstDir, oldFile.Name())
				}
			}
			// find most current source file
			srcFile := get1src(srcDir, srcFiles)
			if len(srcFile) > 0 {
				// compute destination file name
				dstFile := nameDst(dstDir, dstStr, dstAffix)
				// copy
				cp(srcFile, dstFile)
				// remove
				for _, oldFile := range srcFiles {
					rm(srcDir, oldFile.Name())
				}
			} else {
				fmt.Println("NO eligible source files!")
			}
		}
	}
}
