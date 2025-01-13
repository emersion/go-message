package message

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestWriter_multipartWithoutCreatePart(t *testing.T) {
	var h Header
	h.Set("Content-Type", "multipart/alternative; boundary=IMTHEBOUNDARY")

	var b bytes.Buffer
	mw, err := CreateWriter(&b, h)
	if err != nil {
		t.Fatal("Expected no error while creating message writer, got:", err)
	}

	io.WriteString(mw, testMultipartBody)
	mw.Close()

	if s := b.String(); s != testMultipartText {
		t.Errorf("Expected output to be \n%s\n but go \n%s", testMultipartText, s)
	}
}

func TestWriter_multipartWithoutBoundary(t *testing.T) {
	var h Header
	h.Set("Content-Type", "multipart/alternative")

	var b bytes.Buffer
	mw, err := CreateWriter(&b, h)
	if err != nil {
		t.Fatal("Expected no error while creating message writer, got:", err)
	}
	mw.Close()

	e, err := Read(&b)
	if err != nil {
		t.Fatal("Expected no error while reading message, got:", err)
	}

	mediaType, mediaParams, err := e.Header.ContentType()
	if err != nil {
		t.Fatal("Expected no error while parsing Content-Type, got:", err)
	} else if mediaType != "multipart/alternative" {
		t.Errorf("Expected media type to be %q, but got %q", "multipart/alternative", mediaType)
	} else if boundary, ok := mediaParams["boundary"]; !ok || boundary == "" {
		t.Error("Expected boundary to be automatically generated")
	}
}

func TestWriter_afterReading(t *testing.T) {
	original := strings.ReplaceAll(`From: Robin Jarry <robin@jarry.cc>
To: ~rjarry/aerc-devel@lists.sr.ht
Subject: [PATCH aerc] compose: fix deadlock when editor exits with an error
Date: Mon, 13 Jan 2025 14:08:53 +0100
Message-ID: <20250113130852.47802-2-robin@jarry.cc>
MIME-Version: 1.0
Content-Transfer-Encoding: 8bit

When exiting vim with :cq, it exits with an error status which is caught
in the termClosed() callback. This causes the composer tab to be closed
and it is a known and expected behaviour.

Signed-off-by: Robin Jarry <robin@jarry.cc>
---
 app/compose.go | 24 ++++++++++++++++++------
 1 file changed, 18 insertions(+), 6 deletions(-)

diff --git a/app/compose.go b/app/compose.go
index 7a6a423c3ea3..e8a672479245 100644
--- a/app/compose.go
+++ b/app/compose.go
@@ -1197,7 +1197,15 @@ func (c *Composer) reopenEmailFile() error {
 
 func (c *Composer) termClosed(err error) {
 	c.Lock()
-	defer c.Unlock()
+	// RemoveTab() on error must be called *AFTER* c.Unlock() but the defer
+	// statement does the exact opposite (last defer statement is executed
+	// first). Use an explicit list that begins with unlocking first.
+	deferred := []func(){c.Unlock}
+	defer func() {
+		for _, d := range deferred {
+			d()
+		}
+	}()
 	if c.editor == nil {
 		return
 	}
@@ -1205,7 +1213,7 @@ func (c *Composer) termClosed(err error) {
 		PushError("Failed to reopen email file: " + e.Error())
 	}
 	editor := c.editor
-	defer editor.Destroy()
+	deferred = append(deferred, editor.Destroy)
 	c.editor = nil
 	c.focusable = c.focusable[:len(c.focusable)-1]
 	if c.focused >= len(c.focusable) {
@@ -1213,8 +1221,10 @@ func (c *Composer) termClosed(err error) {
 	}
 
 	if editor.cmd.ProcessState.ExitCode() > 0 {
-		RemoveTab(c, true)
-		PushError("Editor exited with error. Compose aborted!")
+		deferred = append(deferred, func() {
+			RemoveTab(c, true)
+			PushError("Editor exited with error. Compose aborted!")
+		})
 		return
 	}
 
@@ -1225,8 +1235,10 @@ func (c *Composer) termClosed(err error) {
 			PushError(err.Error())
 			err := c.showTerminal()
 			if err != nil {
-				RemoveTab(c, true)
-				PushError(err.Error())
+				deferred = append(deferred, func() {
+					RemoveTab(c, true)
+					PushError(err.Error())
+				})
 			}
 			return
 		}
-- 
2.47.1


`, "\n", "\r\n")

	msg, err := Read(strings.NewReader(original))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	msg.Header.Del("Sender")

	var buf bytes.Buffer
	err = msg.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	if buf.String() != original {
		t.Fatalf("reformatted message differs: original=%q reformatted=%q",
			original, buf.String())
	}
}
