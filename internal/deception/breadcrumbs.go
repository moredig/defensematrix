package deception

import (
	"fmt"
	"os"
	"path/filepath"
)

// Breadcrumb represents a fake file dropped inside the Boat
type Breadcrumb struct {
	Filename string
	Content  string
}

// GenerateBreadcrumbs returns a set of fake lure files
func GenerateBreadcrumbs() []Breadcrumb {
	return []Breadcrumb{
		{
			Filename: "credentials.txt",
			Content: `# Internal Credentials - DO NOT SHARE
db_host=192.168.1.100
db_user=admin
db_password=Sup3rS3cr3t!
aws_access_key=AKIAIOSFODNN7EXAMPLE
aws_secret=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`,
		},
		{
			Filename: "backup_partial.sql",
			Content: `-- Database backup (partial)
-- Generated: 2024-01-15 03:00:01
CREATE TABLE users (
  id INT PRIMARY KEY,
  username VARCHAR(50),
  password_hash VARCHAR(255),
  email VARCHAR(100)
);
INSERT INTO users VALUES (1, 'admin', '$2y$10$fakehashedpassword', 'admin@company.internal');
INSERT INTO users VALUES (2, 'root', '$2y$10$anotherfakehash', 'root@company.internal');`,
		},
		{
			Filename: ".env",
			Content: `APP_ENV=production
APP_KEY=base64:fakekey1234567890abcdefghijklmnop
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=production_db
DB_USERNAME=root
DB_PASSWORD=toor
STRIPE_KEY=sk_live_fakestripekey1234567890`,
		},
		{
			Filename: "notes.txt",
			Content: `TODO:
- Move prod DB credentials out of repo (URGENT)
- Rotate AWS keys (done? check with Dave)
- Fix exposed admin panel on port 8080
- Disable root SSH login (pending)`,
		},
	}
}

// DropBreadcrumbs writes fake lure files into a target directory
func DropBreadcrumbs(targetDir string) error {
	crumbs := GenerateBreadcrumbs()
	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return fmt.Errorf("failed to create breadcrumb directory: %w", err)
	}

	for _, crumb := range crumbs {
		path := filepath.Join(targetDir, crumb.Filename)
		if err := os.WriteFile(path, []byte(crumb.Content), 0644); err != nil {
			return fmt.Errorf("failed to drop breadcrumb %s: %w", crumb.Filename, err)
		}
		fmt.Printf("[WAMAI] Breadcrumb dropped: %s\n", crumb.Filename)
	}

	return nil
}
