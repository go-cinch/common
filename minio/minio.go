package minio

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-module/carbon/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Token struct {
	URI  string            `json:"uri,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

type Minio struct {
	ops    Options
	client *minio.Client
}

func New(options ...func(*Options)) (m *Minio, err error) {
	ops := getOptionsOrSetDefault(nil)
	for _, f := range options {
		f(ops)
	}
	if ops.endpoint == "" {
		err = ErrEndpointNil
		return
	}
	if ops.key == "" {
		err = ErrKeyNil
		return
	}
	if ops.secret == "" {
		err = ErrSecretNil
		return
	}
	client, err := minio.New(ops.endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ops.key, ops.secret, ""),
		Secure: ops.ssl,
	})
	if err != nil {
		err = errors.New(strings.Join([]string{ErrInitializeFailed.Error(), err.Error()}, " "))
		return
	}
	if ops.expire != "" {
		_, err = time.ParseDuration(ops.expire)
		if err != nil {
			err = errors.New(strings.Join([]string{ErrExpireInvalid.Error(), err.Error()}, " "))
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// check connection
	bucket := "health"
	if ops.bucket != "" {
		bucket = ops.bucket
	}
	_, err = client.GetBucketLocation(ctx, bucket)
	if err != nil {
		// timeout
		if ctx.Err() != nil {
			err = errors.New(strings.Join([]string{ErrInitializeFailed.Error(), "timeout"}, " "))
			return
		}
		res := minio.ToErrorResponse(err)
		// invalid endpoint/key/secret
		if res.Code != "NoSuchBucket" {
			err = errors.New(strings.Join([]string{ErrInitializeFailed.Error(), err.Error()}, " "))
			return
		}
	}
	m = &Minio{
		ops:    *ops,
		client: client,
	}
	// make bucket if not exist
	if ops.bucket != "" {
		var ok bool
		ok, err = client.BucketExists(ctx, ops.bucket)
		if err != nil {
			return
		}
		if !ok {
			err = client.MakeBucket(ctx, ops.bucket, minio.MakeBucketOptions{
				ObjectLocking: true,
			})
		}
	}
	return
}

func (m *Minio) Token(ctx context.Context, object string) (token Token, err error) {
	if object == "" {
		err = ErrObjectNameNil
		return
	}
	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(m.ops.bucket)
	_ = policy.SetKey(object)
	_ = policy.SetExpires(carbon.Now(carbon.UTC).AddDuration(m.ops.expire).StdTime())
	if m.ops.contentType != "" {
		err = policy.SetContentType(m.ops.contentType)
		if err != nil {
			err = errors.New(strings.Join([]string{ErrContentTypeInvalid.Error(), err.Error()}, " "))
			return
		}
	} else {
		// default any type
		_ = policy.SetContentTypeStartsWith("")
	}
	err = policy.SetContentLengthRange(m.ops.min, m.ops.max)
	if err != nil {
		err = errors.New(strings.Join([]string{ErrObjectSizeInvalid.Error(), err.Error()}, " "))
		return
	}

	uri, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		err = errors.New(strings.Join([]string{ErrTokenSignFailed.Error(), err.Error()}, " "))
		return
	}
	token.URI = uri.String()
	token.Data = formData
	return
}

func (m *Minio) Get(ctx context.Context, object string) (reply string, err error) {
	reader, err := m.client.GetObject(ctx, m.ops.bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return
	}
	defer reader.Close()

	filename := strings.Join([]string{
		m.ops.tmp,
		carbon.Now().ToDateString(),
		filepath.Base(object),
	}, string(filepath.Separator))
	dir := filepath.Dir(filename)

	// create dir if not exist
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return
	}
	// create file
	localFile, err := os.Create(filename)
	if err != nil {
		return
	}
	defer localFile.Close()

	// read file
	stat, err := reader.Stat()
	if err != nil {
		return
	}

	if _, err = io.CopyN(localFile, reader, stat.Size); err != nil {
		return
	}
	reply = filename
	return
}

func (m *Minio) GetObject(ctx context.Context, object string) (*minio.Object, error) {
	return m.client.GetObject(ctx, m.ops.bucket, object, minio.GetObjectOptions{})
}

func (m *Minio) Preview(ctx context.Context, object string) (reply string, err error) {
	duration, _ := time.ParseDuration(m.ops.expire)
	uri, err := m.client.PresignedGetObject(ctx, m.ops.bucket, object, duration, nil)
	if err != nil {
		return
	}
	reply = uri.String()
	return
}

func (m *Minio) FPutObject(ctx context.Context, object, path string) (*minio.UploadInfo, error) {
	info, err := m.client.FPutObject(ctx, m.ops.bucket, object, path, minio.PutObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (m *Minio) PreSignedGetObject(ctx context.Context, object string) (string, error) {
	duration, _ := time.ParseDuration(m.ops.expire)
	res, err := m.client.PresignedGetObject(ctx, m.ops.bucket, object, duration, make(url.Values))
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func (m *Minio) PreSignedPutObject(ctx context.Context, object string) (string, error) {
	duration, _ := time.ParseDuration(m.ops.expire)
	res, err := m.client.PresignedPutObject(ctx, m.ops.bucket, object, duration)
	if err != nil {
		return "", err
	}
	return res.String(), nil
}

func (m *Minio) PreSignedHeadObject(ctx context.Context, object string) (string, error) {
	duration, _ := time.ParseDuration(m.ops.expire)
	res, err := m.client.PresignedHeadObject(ctx, m.ops.bucket, object, duration, make(url.Values))
	if err != nil {
		return "", err
	}
	return res.String(), nil
}
