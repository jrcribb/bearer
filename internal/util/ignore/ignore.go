package ignore

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/exp/maps"

	"github.com/fatih/color"

	"github.com/bearer/bearer/api"
	types "github.com/bearer/bearer/internal/util/ignore/types"
	pointer "github.com/bearer/bearer/internal/util/pointers"
)

func GetIgnoredFingerprints(bearerIgnoreFilePath string, target *string) (ignoredFingerprints map[string]types.IgnoredFingerprint, fileExists bool, err error) {
	if bearerIgnoreFilePath == "" {
		// nothing to do here
		return ignoredFingerprints, false, err
	}

	if target != nil {
		targetPath := ""
		if targetPath, err = filepath.Abs(*target); err != nil {
			return ignoredFingerprints, fileExists, err
		}
		bearerIgnoreFilePath = filepath.Join(targetPath, bearerIgnoreFilePath)
	}

	info, err := os.Stat(bearerIgnoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist : expected scenario
			err = nil
		}
		return make(map[string]types.IgnoredFingerprint), false, err
	}

	if info.IsDir() {
		return ignoredFingerprints, false, fmt.Errorf("bearer-ignore-file path %s is a dir not a file", bearerIgnoreFilePath)
	}

	// file exists
	content, err := os.ReadFile(bearerIgnoreFilePath)
	if err != nil {
		return ignoredFingerprints, true, err
	}

	err = json.Unmarshal(content, &ignoredFingerprints)
	return ignoredFingerprints, true, err
}

func GetIgnoredFingerprintsFromCloud(
	client *api.API,
	fullname string,
	localIgnores map[string]types.IgnoredFingerprint,
) (
	useCloudIgnores bool,
	ignoredFingerprints map[string]types.IgnoredFingerprint,
	staleIgnoredFingerprintIds []string,
	err error,
) {
	data, err := client.FetchIgnores(fullname, maps.Keys(localIgnores))
	if err != nil {
		return useCloudIgnores, ignoredFingerprints, staleIgnoredFingerprintIds, err
	}

	ignoredFingerprints = make(map[string]types.IgnoredFingerprint)
	for _, fingerprint := range data.Ignores {
		ignoredFingerprints[fingerprint] = types.IgnoredFingerprint{}
	}

	return data.ProjectFound, ignoredFingerprints, data.StaleIgnores, nil
}

func MergeIgnoredFingerprints(fingerprintsToIgnore map[string]types.IgnoredFingerprint, ignoredFingerprints map[string]types.IgnoredFingerprint, force bool) error {
	for key, value := range fingerprintsToIgnore {
		if !force {
			if _, ok := ignoredFingerprints[key]; ok {
				return fmt.Errorf(
					"fingerprint '%s' already exists in the bearer.ignore file. To view this entry run:\n\n$ bearer ignore show %s\n\nTo overwrite this entry, use --force",
					key,
					key,
				)
			}
		}
		ignoredAt := time.Now().UTC()
		value.IgnoredAt = ignoredAt.Format(time.RFC3339)
		ignoredFingerprints[key] = value
	}
	return nil
}

var bold = color.New(color.Bold).SprintFunc()
var morePrefix = color.HiBlackString("├─ ")
var lastPrefix = color.HiBlackString("└─ ")

func DisplayIgnoredEntryTextString(fingerprintId string, entry types.IgnoredFingerprint, noColor bool) string {
	initialColorSetting := color.NoColor
	if noColor && !initialColorSetting {
		color.NoColor = true
	}
	prefix := morePrefix
	result := fmt.Sprintf(bold(color.HiBlueString("%s \n")), fingerprintId)

	if entry.Author == nil && entry.Comment == nil {
		prefix = lastPrefix
	}
	result += fmt.Sprintf("%sIgnored At: %s", prefix, bold(entry.IgnoredAt))

	if entry.Author != nil {
		result += fmt.Sprintf("\n%sAuthor: %s", prefix, bold(*entry.Author))
	}

	if entry.Comment == nil {
		prefix = lastPrefix
	}
	var falsePositiveStr string
	if entry.FalsePositive {
		falsePositiveStr = "Yes"
	} else {
		falsePositiveStr = "No"
	}
	result += fmt.Sprintf("\n%sFalse positive? %s", prefix, bold(falsePositiveStr))

	if entry.Comment != nil {
		result += fmt.Sprintf("\n%sComment: %s", lastPrefix, bold(*entry.Comment))
	}

	color.NoColor = initialColorSetting

	return result
}

func GetAuthor() (*string, error) {
	nameBytes, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		return nil, err
	}

	return pointer.String(strings.TrimSuffix(string(nameBytes), "\n")), nil
}