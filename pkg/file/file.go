package file

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func FindFirst(fileSuffix string, dir string) (result string, err error) {
	err = filepath.Walk(dir,
		func(path string, fi os.FileInfo, errIn error) error {
			if strings.HasSuffix(path, fileSuffix) {
				result = path
				return io.EOF
			}
			return nil
		})

	if err == io.EOF {
		err = nil
	}
	return
}

func FindAll(suffix string, excludes []string, dir string) (result []string, err error) {
	err = filepath.Walk(dir,
		func(path string, fi os.FileInfo, errIn error) error {
			if strings.HasSuffix(path, suffix) && !SuffixIn(path, excludes) {
				result = append(result, path)
			}
			return nil
		})

	if err == io.EOF {
		err = nil
	}
	return
}

func SuffixIn(keyword string, list []string) bool {
	for _, w := range list {
		if strings.Contains(keyword, w) {
			return true
		}
	}
	return false
}

func ReadJson(file string, parsed interface{}) error {
	byteValue, err := Open(file)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to read %s, %v", file, err))
	}

	err = json.Unmarshal(byteValue, &parsed)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to unmarshal %s, %v", file, err))
	}

	return nil
}

func ReadXml(file string, parsed interface{}) error {
	byteValue, err := Open(file)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to open %s, %v", file, err))
	}

	err = xml.Unmarshal(byteValue, &parsed)
	if err != nil {
		return errors.New(fmt.Sprintf("Unable to unmarshal %s, %v", file, err))
	}

	return nil
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func Open(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return []byte{}, err
	}

	defer file.Close()

	return byteValue, nil
}

func OpenLines(filePath string) ([]string, error) {
	b, err := Open(filePath)
	if err != nil {
		return []string{}, nil
	}

	return strings.Split(string(b), "\n"), nil
}

func Overwrite(lines []string, filePath string) error {
	return ioutil.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0644)
}

func CopyOrMerge(sourceFile string, destinationFile string) error {
	if Exists(destinationFile) {
		return mergeFile(sourceFile, destinationFile)
	}

	return copyFile(sourceFile, destinationFile)
}

func mergeFile(sourceFile string, destinationFile string) error {
	if strings.HasSuffix(sourceFile, ".java") || strings.HasSuffix(sourceFile, ".kt") {
		log.Infof(fmt.Sprintf("ignoring merge of java or kt files: %s", sourceFile))
		return nil
	}

	if strings.HasSuffix(sourceFile, ".properties") {
		log.Infof("merging key=val property file %s with %s", sourceFile, destinationFile)
		return MergeKeyValFile(sourceFile, destinationFile, "=")
	}

	log.Infof("merging text file %s with %s", sourceFile, destinationFile)
	return MergeTextFiles(sourceFile, destinationFile)
}

func copyFile(sourceFile string, destinationFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}

	pathSeparator := string(os.PathSeparator)
	destinationParts := strings.Split(destinationFile, pathSeparator)
	destinationDir := strings.Join(destinationParts[:len(destinationParts)-1], pathSeparator)
	if !Exists(destinationDir) {
		err = CreateDirectory(destinationDir)
		if err != nil {
			return err
		}
	}

	fileInfo, err := os.Stat(sourceFile)
	if err != nil {
		return err
	}

	log.Infof("copying FROM\t <= %s", sourceFile)
	log.Infof("copying TO\t => %s", destinationFile)
	err = ioutil.WriteFile(destinationFile, input, fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func RelPath(sourceDirectory string, filePath string) (string, error) {
	pathSeparator := string(os.PathSeparator)
	directoryParts := strings.Split(sourceDirectory, pathSeparator)
	fileParts := strings.Split(filePath, pathSeparator)

	if len(directoryParts) >= len(fileParts) {
		return "", errors.New("directory cannot be deeper than filePath")
	}

	cut := 0

	for i := range directoryParts {
		if directoryParts[i] == fileParts[i] {
			cut += 1
		} else {
			break
		}
	}

	return strings.Join(fileParts[cut:], pathSeparator), nil
}

func CreateDirectory(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			return err
		}
	}

	return nil
}

func SearchReplace(filePath string, from string, to string) error {
	b, err := Open(filePath)
	if err != nil {
		return err
	}

	replaced := strings.ReplaceAll(string(b), from, to)
	return Overwrite(strings.Split(replaced, "\n"), filePath)
}

func MergeKeyValFile(fromFile string, toFile string, separator string) error {
	fromLines, err := OpenLines(fromFile)
	if err != nil {
		return err
	}

	toLines, err := OpenLines(toFile)
	if err != nil {
		return err
	}

	var newLines []string
	for _, fromLine := range fromLines {
		if fromLine == "" || !strings.Contains(fromLine, separator) {
			continue
		}

		var hasLine = false
		fromParts := strings.Split(fromLine, separator)
		fromKey := fromParts[0]

		for _, toLine := range toLines {
			if toLine == "" && !strings.Contains(toLine, separator) {
				continue
			}
			toParts := strings.Split(toLine, separator)
			toKey := toParts[0]
			if fromKey == toKey {
				log.Debugf("ignoring line due to key duplicate found in source %s: '%s' and '%s' in target:%s", fromFile, fromLine, toLine, toFile)
				hasLine = true
			}
		}
		if !hasLine {
			newLines = append(newLines, fromLine)
			log.Debugf("appending line: '%s', to:%s", fromLine, toFile)
		}
	}

	toLines = append(toLines, newLines...)

	return Overwrite(toLines, toFile)
}

func MergeTextFiles(fromFile string, toFile string) error {
	fromLines, err := OpenLines(fromFile)
	if err != nil {
		return err
	}

	toLines, err := OpenLines(toFile)
	if err != nil {
		return err
	}

	var newLines []string
	for _, fromLine := range fromLines {
		var hasLine = false
		for _, toLine := range toLines {
			if fromLine == toLine {
				hasLine = true
			}
		}
		if !hasLine {
			newLines = append(newLines, fromLine)
			log.Debugf("appending line: '%s', to:%s", fromLine, toFile)
		}
	}

	toLines = append(toLines, newLines...)

	return Overwrite(toLines, toFile)
}

func Equal(fileA string, fileB string) (bool, error) {
	fileALines, err := OpenLines(fileA)
	if err != nil {
		return false, err
	}
	fileBLines, err := OpenLines(fileB)
	if err != nil {
		return false, err
	}

	if len(fileALines) != len(fileBLines) {
		return false, nil
	}

	for i := range fileALines {
		if fileALines[i] != fileBLines[i] {
			return false, nil
		}
	}

	return true, nil
}

func DeleteSingleFile(filePath string) error {
	return os.Remove(filePath)
}

func DeleteAll(dirPath string) error {
	return os.RemoveAll(dirPath)
}
