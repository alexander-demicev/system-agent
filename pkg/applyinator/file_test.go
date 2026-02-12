package applyinator

import (
	"encoding/base64"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("File Operations", func() {
	var tempDir string

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Describe("WriteContentToFile", func() {
		var getFile = func(path string, permissions string, content string) File {
			return File{
				Content:     content,
				UID:         -1,
				GID:         -1,
				Path:        filepath.Join(tempDir, path),
				Permissions: permissions,
			}
		}

		Context("with no permissions specified", func() {
			It("should create file with default permissions", func() {
				f := getFile("test-no-perms", "", "hello world")
				perms, err := parsePerm(f.Permissions)
				Expect(err).To(HaveOccurred())
				err = writeContentToFile(f.Path, f.UID, f.GID, perms, []byte(f.Content))
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(f.Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(content).To(Equal([]byte("hello world")))

				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(defaultFilePermissions))
				}
			})
		})

		Context("with custom permissions", func() {
			It("should create file with specified permissions", func() {
				f := getFile("test-perms", "0666", "hello world 2")
				perms, err := parsePerm(f.Permissions)
				Expect(err).ToNot(HaveOccurred())
				err = writeContentToFile(f.Path, f.UID, f.GID, perms, []byte(f.Content))
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(f.Path)
				Expect(err).ToNot(HaveOccurred())
				Expect(content).To(Equal([]byte("hello world 2")))
				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(os.FileMode(0666)))
				}
			})
		})

		Context("with base64 encoded content", func() {
			It("should fail with invalid base64", func() {
				f := getFile("test-invalid-base64", "", "not base64 content")
				err := writeBase64ContentToFile(f)
				Expect(err).To(HaveOccurred())
			})

			It("should decode and write base64 content with default permissions", func() {
				f := getFile("test-no-perms-base64", "", "aGVsbG8gd29ybGQ=")
				err := writeBase64ContentToFile(f)
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(f.Path)
				Expect(err).ToNot(HaveOccurred())
				decoded, err := base64.StdEncoding.DecodeString("aGVsbG8gd29ybGQ=")
				Expect(err).ToNot(HaveOccurred())
				Expect(content).To(Equal(decoded))

				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(defaultFilePermissions))
				}
			})

			It("should decode and write base64 content with custom permissions", func() {
				f := getFile("test-perms-base64", "0666", "aGVsbG8gd29ybGQ=")
				err := writeBase64ContentToFile(f)
				Expect(err).ToNot(HaveOccurred())

				content, err := os.ReadFile(f.Path)
				Expect(err).ToNot(HaveOccurred())
				decoded, err := base64.StdEncoding.DecodeString("aGVsbG8gd29ybGQ=")
				Expect(err).ToNot(HaveOccurred())
				Expect(content).To(Equal(decoded))

				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(os.FileMode(0666)))
				}
			})
		})
	})

	Describe("CreateDirectory", func() {
		var getFile = func(path string, permissions string) File {
			return File{
				Directory:   true,
				UID:         -1,
				GID:         -1,
				Path:        filepath.Join(tempDir, path),
				Permissions: permissions,
			}
		}

		Context("with no permissions specified", func() {
			It("should create directory with default permissions", func() {
				f := getFile("test-no-perms", "")
				err := createDirectory(f)
				Expect(err).ToNot(HaveOccurred())

				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(fs.ModeDir | defaultDirectoryPermissions))
				}
			})
		})

		Context("with custom permissions", func() {
			It("should create directory with specified permissions", func() {
				f := getFile("test-perms", "0777")
				err := createDirectory(f)
				Expect(err).ToNot(HaveOccurred())

				if runtime.GOOS != "windows" {
					permissions, err := getPermissions(f.Path)
					Expect(err).ToNot(HaveOccurred())
					Expect(permissions).To(Equal(fs.ModeDir | os.FileMode(0777)))
				}
			})
		})
	})

	Describe("ParsePerm", func() {
		DescribeTable("parsing permission strings",
			func(permissions string, expectedPerms os.FileMode, expectErr bool) {
				fileMode, err := parsePerm(permissions)
				if expectErr {
					Expect(err).To(HaveOccurred())
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(fileMode).To(Equal(expectedPerms))
				}
			},
			Entry("should parse 0777", "0777", os.FileMode(0777), false),
			Entry("should parse 0007", "0007", os.FileMode(0007), false),
			Entry("should parse 0070", "0070", os.FileMode(0070), false),
			Entry("should parse 0700", "0700", os.FileMode(0700), false),
			Entry("should parse 0333", "0333", os.FileMode(0333), false),
			Entry("should parse 0003", "0003", os.FileMode(0003), false),
			Entry("should parse 0030", "0030", os.FileMode(0030), false),
			Entry("should parse 0300", "0300", os.FileMode(0300), false),
			Entry("should error on empty string", "", os.FileMode(0), true),
		)
	})

	Describe("FileActionDelete", func() {
		var getFile = func(path string, isDir bool) File {
			return File{
				Path:      filepath.Join(tempDir, path),
				Directory: isDir,
				Action:    deleteFileAction,
			}
		}

		Context("when deleting existing directory", func() {
			It("should successfully remove the directory", func() {
				dirPath := filepath.Join(tempDir, "existing-dir")
				err := os.Mkdir(dirPath, defaultDirectoryPermissions)
				Expect(err).ToNot(HaveOccurred())

				f := getFile("existing-dir", true)
				err = removeFile(f)
				Expect(err).ToNot(HaveOccurred())

				_, err = os.Stat(dirPath)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})

		Context("when deleting missing directory", func() {
			It("should not return error", func() {
				f := getFile("missing-dir", true)
				err := removeFile(f)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when deleting existing file", func() {
			It("should successfully remove the file", func() {
				filePath := filepath.Join(tempDir, "existing-file")
				err := os.WriteFile(filePath, []byte("t"), defaultFilePermissions)
				Expect(err).ToNot(HaveOccurred())

				f := getFile("existing-file", false)
				err = removeFile(f)
				Expect(err).ToNot(HaveOccurred())

				_, err = os.Stat(filePath)
				Expect(os.IsNotExist(err)).To(BeTrue())
			})
		})

		Context("when deleting missing file", func() {
			It("should not return error", func() {
				f := getFile("missing-file", false)
				err := removeFile(f)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
