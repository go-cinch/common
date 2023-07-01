package minio

type Options struct {
	endpoint    string // minio node endpoint
	key         string // minio access key id
	secret      string // minio access secret
	bucket      string // default bucket name
	ssl         bool   // use https or not
	expire      string // Token expire time
	contentType string // allow upload content type
	min         int64  // allow upload min size
	max         int64  // allow upload max size
	tmp         string // Get file save path
}

func WithEndpoint(endpoint string) func(*Options) {
	return func(options *Options) {
		if endpoint != "" {
			getOptionsOrSetDefault(options).endpoint = endpoint
		}
	}
}

func WithKey(key string) func(*Options) {
	return func(options *Options) {
		if key != "" {
			getOptionsOrSetDefault(options).key = key
		}
	}
}

func WithSecret(secret string) func(*Options) {
	return func(options *Options) {
		if secret != "" {
			getOptionsOrSetDefault(options).secret = secret
		}
	}
}

func WithBucket(bucket string) func(*Options) {
	return func(options *Options) {
		if bucket != "" {
			getOptionsOrSetDefault(options).bucket = bucket
		}
	}
}

func WithSSL(ssl bool) func(*Options) {
	return func(options *Options) {
		getOptionsOrSetDefault(options).ssl = ssl
	}
}

func WithExpire(expire string) func(*Options) {
	return func(options *Options) {
		if expire != "" {
			getOptionsOrSetDefault(options).expire = expire
		}
	}
}

func WithContentType(contentType string) func(*Options) {
	return func(options *Options) {
		if contentType != "" {
			getOptionsOrSetDefault(options).contentType = contentType
		}
	}
}

func WithMin(min int64) func(*Options) {
	return func(options *Options) {
		if min > 0 {
			getOptionsOrSetDefault(options).min = min
		}
	}
}

func WithMax(max int64) func(*Options) {
	return func(options *Options) {
		if max > 0 {
			getOptionsOrSetDefault(options).max = max
		}
	}
}

func WithTmp(tmp string) func(*Options) {
	return func(options *Options) {
		if tmp != "" {
			getOptionsOrSetDefault(options).tmp = tmp
		}
	}
}

func getOptionsOrSetDefault(options *Options) *Options {
	if options == nil {
		return &Options{
			expire: "24h",             // one day, ns/us/ms/s/m/h
			min:    1024,              // 1KB
			max:    1024 * 1024 * 500, // 500MB
			tmp:    "/tmp",
		}
	}
	return options
}
