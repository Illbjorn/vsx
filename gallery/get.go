package gallery

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
)

type VoltronReader interface {
	io.Reader
	io.ReaderAt
	io.ByteReader
	io.RuneReader
	io.Seeker
}

// GetExtension accepts a gallery publisherID, extension ID and version
// returning an io.ReadCloser containing the binary payload of the requested
// extension's VSIX package.
func (self Gallery) GetExtension(
	ctx context.Context,
	publisherID, extensionID, version string,
) (VoltronReader, error) {
	const assetKindVSIXPackage = "Microsoft.VisualStudio.Services.VSIXPackage"
	const pathFmtGetExtension = "_apis/public/gallery/publisher/" +
		"%s" /* [1] Publisher ID      */ + "/extension/" +
		"%s" /* [2] Extension ID      */ + "/" +
		"%s" /* [3] Extension Version */ + "/assetbyname/" + assetKindVSIXPackage

	// Construct the URL
	path := fmt.Sprintf(pathFmtGetExtension, publisherID, extensionID, version)
	url := self.BaseURL.JoinPath(path)

	// Init the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init GET request: %w", err)
	}

	// Get the response
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute GET request to [%s]: %w",
			url.String(), err,
		)
	}
	defer res.Body.Close()

	// The response body is the VSIX package, we need this as an `io.ReaderAt`
	// to pass it to `zip.NewReader()`. To achieve this we read the response body
	// and wrap it back into a `bytes.Reader`.

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read extension response body: %w", err)
	}

	// Evaluate request failures
	//
	// We include the response body in the error message if the status code is
	// >= 400 (hence this conditional being >1 step from the actual doing of the
	// request)
	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf(
			"received received HTTP status code [%d] in GET request to [%s]: %s",
			res.StatusCode, url.String(), string(body),
		)
	}

	// Wrap into `bytes.Reader` which implements `io.ReaderAt`
	r := bytes.NewReader(body)

	// https://i.imgflip.com/5g7vmt.jpg
	return r, nil
}
