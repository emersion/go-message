package message

import (
	"reflect"
	"testing"
)

func TestHeader(t *testing.T) {
	mediaType := "multipart/mixed"
	params := map[string]string{"boundary": "my-boundary"}
	desc := "Plan de complémentarité de l'Homme"

	h := make(Header)
	h.SetContentType(mediaType, params)
	h.SetContentDescription(desc)

	if gotMediaType, gotParams, err := h.ContentType(); err != nil {
		t.Error("Expected no error when parsing content type, but got:", err)
	} else if gotMediaType != mediaType {
		t.Errorf("Expected media type %q but got %q", mediaType, gotMediaType)
	} else if !reflect.DeepEqual(gotParams, params) {
		t.Errorf("Expected media params %v but got %v", params, gotParams)
	}

	if gotDesc, err := h.ContentDescription(); err != nil {
		t.Error("Expected no error when parsing content description, but got:", err)
	} else if gotDesc != desc {
		t.Errorf("Expected content description %q but got %q", desc, gotDesc)
	}
}
