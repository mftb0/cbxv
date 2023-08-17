package util

import (
	"archive/zip"
	"crypto/md5"
	"embed"
    "errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gen2brain/go-unarr"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

const RENDERERSTATE_FN string = "rendererstate.json"
const CBXS_DN string = "cbxv"
const BOOKMARKS_DN string = "bookmarks"
const LAYOUTS_DN string = "layouts"
const TMP_CBXS_PREFIX string = "cbxv-"
const DEBUG = false

const HELP_TXT = `<tt>
Command             Key                 Mouse
quit                q                   Window Close Button 
openFile            o                   File Button
closeFile           c                   NA
rightPage           d|l|[RightArrow]    Right Side
leftPage            a|h|[LeftArrow]     Left Side
firstPage           w|k|[UpArrow]       NA
lastPage            s|j|[DownArrow]     NA
nextFile            n                   NA
previousFile        p                   NA
toggleBookmark      [Space]             Bookmark Buttons
lastBookmark        L                   NA
help                ?|[F1]              Question Mark Button
toggleDirection     [BackTick]          Direction Toggle
1-Page Layout       1                   NA
2-Page Layout       2                   NA
stripLayout         3                   NA
hidePage            -                   NA
toggleJoin          r                   Join Toggle
toggleFullscreen    f|[F11]             Fullscreen Toggle
selectPage          [Tab]               Page Index Buttons
exportPage          e                   NA

<a href="https://mftb0.github.io/cbxv">Program Manual</a>

Copyright © 2022-2023 Matt Butler
</tt>`

//go:embed assets
var assets embed.FS

// Stuff to handle messages - model -> app <- ui
type Message struct {
    TypeName string
    Data     string
}

type Messenger func(m Message)

func tmpPath() string {
    return os.TempDir()
}

func homePath() (string, error) {
    return os.UserHomeDir()
}

func configPath() (string, error) {
    p, err := os.UserConfigDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(p, CBXS_DN), nil
}

func cachePath() (string, error) {
    p, err := os.UserCacheDir()
    if err != nil {
        return "", err
    }
    return filepath.Join(p, CBXS_DN), nil
}

func dataPath() (string, error) {
    p, err := homePath()
    if err != nil {
        return "", err
    }

    dataHome := filepath.Join(".local", "share")
    if strings.Contains(runtime.GOOS, "windows") {
        dataHome = filepath.Join("AppData", "Roaming")
    } else if strings.Contains(runtime.GOOS, "darwin") {
        dataHome = filepath.Join("Library", "Application Support")
    }

    return filepath.Join(p, dataHome, CBXS_DN), nil
}

func rendererstatePath() (string, error) {
    p, err := cachePath()
    if err != nil {
        return p, err
    }
    return filepath.Join(p, RENDERERSTATE_FN), nil
}

func bookmarksPath() (string, error) {
    p, err := dataPath()
    if err != nil {
        return p, err
    }
    return filepath.Join(p, BOOKMARKS_DN), nil
}

func layoutsPath() (string, error) {
    p, err := dataPath()
    if err != nil {
        return p, err
    }
    return filepath.Join(p, LAYOUTS_DN), nil
}

// Get a string that points to an icon for our executables use at runtime
// Linux - find the icon in one of the standard directories
// Windows - find the icon relative to the executable
func AppIconPath() (*string) {
    var p string
    fname := "logo_cbxv.png"
    if runtime.GOOS == "linux" {
        sharePath := "/share/icons/hicolor/1024x1024/apps"

        hPath, err := homePath()
        if err != nil {
            return nil
        }

        iPath := filepath.Join(hPath, ".local", sharePath, fname)
        if _, err := os.Stat(iPath); errors.Is(err, os.ErrNotExist) {
            iPath := filepath.Join("/usr", sharePath, fname)
            if _, err := os.Stat(iPath); err == nil {
                p = iPath
            }
        } else {
            p = iPath
        }
    } else if runtime.GOOS == "windows" {
        ePath, err := os.Executable()
        if err != nil {
            return nil
        }
        eDir := filepath.Dir(ePath)
        p = filepath.Join(eDir, fname)
    }

    if len(p) > 0 {
        return &p
    }
    return nil
}

func FullscreenIcon() string {
    // "⛶ "  u+26f6 - square four corners
    // "⤢ "  u+2922 - NE/SW Arrows
    // "[ ]" Regular square brackets

    r := "⛶ " // u+26f6 - square four corners
    if strings.Contains(runtime.GOOS, "linux") {
        //noop
    } else if strings.Contains(runtime.GOOS, "windows") {
        //noop
    } else if strings.Contains(runtime.GOOS, "darwin") {
        r = "[ ]" // Regular square brackets
    } else {
        r = "[ ]" // Regular square brackets
    }

    return r
}

func RestoreIcon() string {
    r := "🮻 " // u+1fbbb - voided greek cross
    if strings.Contains(runtime.GOOS, "linux") {
        //noop
    } else if strings.Contains(runtime.GOOS, "windows") {
        r = "╬" // line-drawing voided cross
    } else if strings.Contains(runtime.GOOS, "darwin") {
        r = "╬" // line-drawing voided cross
    } else {
        r = "╬" // line-drawing voided cross
    }

    return r
}

func createRandomString(n int) string {
    var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
    rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letter[rand.Intn(len(letter))]
    }
    return string(b)
}

func validCBXExt(ext string) bool {
     if ext != ".cbz" &&
        ext != ".cbr" &&
        ext != ".cb7" &&
        ext != ".cbt" &&
        ext != ".pdf" {
        return false
    }
    return true
}

func validImageExt(ext string) bool {
    if ext != ".jpg" &&
        ext != ".jpeg" &&
        ext != ".png" &&
        ext != ".webp" &&
        ext != ".avif" &&
        ext != ".heic" &&
        ext != ".gif" {
        return false
    }
    return true
}

func extractZipWorker(dst *os.File, src io.ReadCloser, errors chan error) {
    _, err := io.Copy(dst, src)
    dst.Close()
    src.Close()
    if err != nil {
        errors <- err
    } else {
        errors <- nil
    }
}

func extractZip(filePath string, tmpDir string) ([]string, error) {

    r, err := zip.OpenReader(filePath)
    if err != nil {
        return nil, err
    }
    defer r.Close()

    var urls []string
    workerMax := runtime.NumCPU()
    workerCount := 0
    workerErrors := make(chan error, workerMax)
    for _, f := range r.File {
        fp := filepath.Join(tmpDir, f.Name)

        // If entry is a dir create and we're done
        if f.FileInfo().IsDir() {
            os.MkdirAll(fp, os.ModePerm)
            continue
        }

        // If extension isn't a useful one we're done
        ext := strings.ToLower(filepath.Ext(fp))
        if !validImageExt(ext) {
            continue
        }

        url := fp
        urls = append(urls, url)

        if err = os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
            return nil, err
        }

        // Closed in extractZipWorker
        dst, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            return nil, err
        }

        // Closed in extractZipWorker
        entry, err := f.Open()
        if err != nil {
            return nil, err
        }

        // Create workers up to num cpus
        if workerCount < workerMax {
            workerCount++

            // A worker extracts a file from the zip and
            // writes errs to errs channel
            go extractZipWorker(dst, entry, workerErrors)
        } else {

            // Loop blocking for workers to report
            // We bail on the first err
            for i := 0; i < workerMax; i++ {
                err := <-workerErrors
                workerCount--
                if err != nil {
                    return nil, err
                }
            }

            // Now that we've cleaned up, handle this zip entry
            workerCount++
            go extractZipWorker(dst, entry, workerErrors)
        }
    }

    // Wait for stragglers
    for i := workerCount; i > 0; i-- {
        err := <-workerErrors
        if err != nil {
            return nil, err
        }
    }

    sort.Strings(urls)
    return urls, nil
}

func extractPdf(filePath string, tmpDir string) ([]string, error) {
    err := os.MkdirAll(tmpDir, os.ModePerm)
    if err != nil {
        return nil, err 
    }

    err = api.ExtractImagesFile(filePath, tmpDir, nil, nil)
    if err != nil {
        return nil, err 
    }

    urls := make([]string, 0)
    entries, err := os.ReadDir(tmpDir)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        ext := strings.ToLower(filepath.Ext(entry.Name()))
        if !validImageExt(ext) {
            continue
        }
        entryPath := filepath.Join(tmpDir, entry.Name())
        urls = append(urls, entryPath)
    }

    sort.Strings(urls)
    return urls, nil
}

func extractRar(filePath string, tmpDir string) ([]string, error) {
    a, err := unarr.NewArchive(filePath)
    if err != nil {
        return nil, err
    }
    defer a.Close()
    entries, err := a.Extract(tmpDir)
    if err != nil {
        return nil, err
    }

    var urls []string
    for _, entry := range entries {
        ext := strings.ToLower(filepath.Ext(entry))
        if !validImageExt(ext) {
            continue
        }

        fp := filepath.Join(tmpDir, entry)
        url := fp
        urls = append(urls, url)
    }

    sort.Strings(urls)
    return urls, nil
}

func extract(filePath string, tmpDir string) ([]string, error) {

    ext := filepath.Ext(filePath)
    if ext == ".pdf" {
        result, err := extractPdf(filePath, tmpDir)
        if err != nil {
            return nil, err
        } else {
           return result, nil
        }
    }

    result, err := extractZip(filePath, tmpDir)
    if err != nil {
        rr, err := extractRar(filePath, tmpDir)
        if err != nil {
            return nil, err
        }
        result = rr
    }
    return result, nil
}

func isDirLink(entry os.DirEntry, filepath string) bool {
    info, err := entry.Info()
    if err != nil {
        return false
    }

    if info.Mode()&fs.ModeSymlink != 0 {
        dst, err := os.Readlink(filepath)
        if err != nil {
            return false
        }

        dstInfo, err := os.Stat(dst)
        if err != nil {
            return false
        }

        if dstInfo.IsDir() {
            return true
        }
    }
    return false
}

func rmCBXTmpDir(dirname string) error {
    tmpDir := filepath.Join(tmpPath(), dirname)
    return os.RemoveAll(tmpDir)
}

func CreateTmpDir() (string, error) {
    rs := createRandomString(6)
    tp := filepath.Join(tmpPath(), fmt.Sprintf("%s%s", TMP_CBXS_PREFIX, rs))
    return tp, nil
}

func GetImagePaths(filePath string, tmpDir string) ([]string, error) {
    return extract(filePath, tmpDir)
}

func ExportPage(srcPath string, dstPath string) error {
    src, err := os.Open(srcPath)
    if err != nil {
        return err
    }
    defer src.Close()

    dst, err := os.Create(dstPath)
    if err != nil {
        return err
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    if err != nil {
        return err
    }
    return nil
}

func HashFile(filePath string) (string, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return "", err
    }
    defer f.Close()

    h := md5.New()
    if _, err := io.Copy(h, f); err != nil {
        return "", err
    }

    hash := fmt.Sprintf("%x", h.Sum(nil))
    return hash, nil
}

func WriteBookmarkList(hash string, data string) error {
    bPath, err := bookmarksPath()
    if err != nil {
        return err
    }
    if err := os.MkdirAll(bPath, 0777); err != nil {
        return err
    }

    storePath := filepath.Join(bPath, fmt.Sprintf("%s.json", hash))
    os.WriteFile(storePath, []byte(data), 0777)
    return nil
}

func ReadBookmarkList(hash string) (*string, error) {
    bkmarksPath, err := bookmarksPath()
    if err != nil {
        return nil, err
    }

    bkmarksPath = filepath.Join(bkmarksPath, hash)
    bkmarksPath = fmt.Sprintf("%s.json", bkmarksPath)

    b, err := ioutil.ReadFile(bkmarksPath)
    if err != nil {
        return nil, err
    }
    s := string(b)
    return &s, nil
}

func WriteLayout(hash string, data string) error {
    lPath, err := layoutsPath()
    if err != nil {
        return err
    }
    if err := os.MkdirAll(lPath, 0777); err != nil {
        return err
    }

    storePath := filepath.Join(lPath, fmt.Sprintf("%s.json", hash))
    os.WriteFile(storePath, []byte(data), 0777)
    return nil
}

func ReadLayout(hash string) (*string, error) {
    lPath, err := layoutsPath()
    if err != nil {
        return nil, err
    }

    lPath = filepath.Join(lPath, hash)
    lPath = fmt.Sprintf("%s.json", lPath)

    b, err := ioutil.ReadFile(lPath)
    if err != nil {
        return nil, err
    }
    s := string(b)
    return &s, nil
}

func WriteRendererState(data string) error {
    cPath, err := configPath()
    if err != nil {
        return err
    }

    if err := os.MkdirAll(cPath, os.ModeDir); err != nil {
        return err
    }

    storePath, err := rendererstatePath()
    if err != nil {
        return err
    }

    return os.WriteFile(storePath, []byte(data), 0666)
}

func ReadRendererState() (string, error) {
    fn, err := rendererstatePath()
    if err != nil {
        return "", err
    }

    s, err := ioutil.ReadFile(fn)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s", s), nil
}

func ReadSeriesList(filePath string) ([]string, error) {
    dirname := filepath.Dir(filePath)
    list := make([]string, 0)
    entries, err := os.ReadDir(dirname)
    if err != nil {
        return nil, err
    }

    for _, entry := range entries {
        ext := strings.ToLower(filepath.Ext(entry.Name()))
        if !validCBXExt(ext){
            continue
        }
        entryPath := filepath.Join(dirname, entry.Name())
        list = append(list, entryPath)
    }
    sort.Strings(list)
    return list, nil
}

type DirListItem struct {
    Item_path string `json:"item_path"`
    Item_type string `json:"item_type"`
}

func ReadDirList(filePath string) ([]DirListItem, error) {
    var err error
    if filePath == "" {
        filePath, err = os.Getwd()
        if err != nil {
            return nil, err
        }
    }

    list := make([]DirListItem, 0)
    entries, err := os.ReadDir(filePath)
    for _, entry := range entries {
        var item DirListItem
        entryPath := filepath.Join(filePath, entry.Name())
        if entry.IsDir() {
            item = DirListItem{
                entryPath,
                "directory",
            }
        } else {
            if isDirLink(entry, entryPath) {
                item = DirListItem{
                    entryPath,
                    "directory",
                }
            } else {
                item = DirListItem{
                    entryPath,
                    "file",
                }
            }
        }
        list = append(list, item)
    }
    return list, nil
}

func LoadTextFile(filePath string) (*string, error) {
    b, err := assets.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    s := string(b)
    return &s, nil
}

func S(str string) *string { return &str }

func ParseFileUrl(fileUrl string) *string {
    var r string
    if strings.HasPrefix(fileUrl, "file:///") {
        uri, err := url.ParseRequestURI(strings.Trim(fileUrl, "\r\n\t"))
        if err != nil {
            fmt.Printf("Error parsing file url - %s", err)
            return &r
        }

        if uri != nil {
            // For windows a forward slash precedes the drive leter o_0
            if runtime.GOOS == "windows" {
                r = uri.Path[1:]
            } else {
                r = uri.Path
            }
        }
    }
    return &r
}

func Log(format string, a ...any) {
    if DEBUG == true {
        fmt.Fprintf(os.Stdout, format, a...)
    }
}

func TrcStart (s string) (string, time.Time) {
    return s, time.Now() 
}

func TrcEnd (s string, startTime time.Time) {
    endTime := time.Now() 
    fmt.Printf("Name:%s, Time:%d\n", s, endTime.Sub(startTime).Milliseconds())
}

