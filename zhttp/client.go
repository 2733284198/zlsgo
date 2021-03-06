/*
 * @Author: seekwe
 * @Date:   2019-05-30 13:19:45
 * @Last Modified by:   seekwe
 * @Last Modified time: 2020-01-31 22:44:40
 */

package zhttp

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/sohaha/zlsgo/zlog"
	"github.com/sohaha/zlsgo/zstring"
)

func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   2 * time.Minute,
	}
}

func (r *Engine) Client() *http.Client {
	if r.client == nil {
		r.client = newClient()
	}
	return r.client
}

func (r *Engine) SetClient(client *http.Client) {
	r.client = client
}

func (r *Engine) Get(url string, v ...interface{}) (*Res, error) {
	return r.Do("GET", url, v...)
}

func (r *Engine) Post(url string, v ...interface{}) (*Res, error) {
	return r.Do("POST", url, v...)
}

func (r *Engine) Put(url string, v ...interface{}) (*Res, error) {
	return r.Do("PUT", url, v...)
}

func (r *Engine) Patch(url string, v ...interface{}) (*Res, error) {
	return r.Do("PATCH", url, v...)
}

func (r *Engine) Delete(url string, v ...interface{}) (*Res, error) {
	return r.Do("DELETE", url, v...)
}

func (r *Engine) Head(url string, v ...interface{}) (*Res, error) {
	return r.Do("HEAD", url, v...)
}

func (r *Engine) Options(url string, v ...interface{}) (*Res, error) {
	return r.Do("OPTIONS", url, v...)
}

func (r *Engine) EnableInsecureTLS(enable bool) {
	trans := r.getTransport()
	if trans == nil {
		return
	}
	if trans.TLSClientConfig == nil {
		trans.TLSClientConfig = &tls.Config{}
	}
	trans.TLSClientConfig.InsecureSkipVerify = enable
}

func (r *Engine) EnableCookie(enable bool) {
	if enable {
		jar, _ := cookiejar.New(nil)
		r.Client().Jar = jar
	} else {
		r.Client().Jar = nil
	}
}

func (r *Engine) CheckRedirect(fn ...func(req *http.Request, via []*http.Request) error) {
	if len(fn) > 0 {
		r.Client().CheckRedirect = fn[0]
	} else {
		r.Client().CheckRedirect = func(_ *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
}

func (r *Engine) SetTimeout(d time.Duration) {
	r.Client().Timeout = d
}

func (r *Engine) SetProxyUrl(proxyUrl ...string) error {
	proxylen := len(proxyUrl)
	if proxylen == 0 {
		return errors.New("proxyurl cannot be empty")
	}
	rawurl := proxyUrl[0]
	if proxylen > 1 {
		rawurl = proxyUrl[zstring.RandInt(0, proxylen-1)]
	}
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	trans.Proxy = http.ProxyURL(u)
	return nil
}

func (r *Engine) SetProxy(proxy func(*http.Request) (*url.URL, error)) error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	trans.Proxy = proxy
	return nil
}

func (r *Engine) RemoveProxy() error {
	trans := r.getTransport()
	if trans == nil {
		return ErrNoTransport
	}
	trans.Proxy = http.ProxyFromEnvironment
	return nil
}

func (r *Engine) getJSONEncOpts() *jsonEncOpts {
	if r.jsonEncOpts == nil {
		r.jsonEncOpts = &jsonEncOpts{escapeHTML: true}
	}
	return r.jsonEncOpts
}

func (r *Engine) SetJSONEscapeHTML(escape bool) {
	opts := r.getJSONEncOpts()
	opts.escapeHTML = escape
}

func (r *Engine) SetJSONIndent(prefix, indent string) {
	opts := r.getJSONEncOpts()
	opts.indentPrefix = prefix
	opts.indentValue = indent
}

func (r *Engine) SetXMLIndent(prefix, indent string) {
	opts := r.getXMLEncOpts()
	opts.prefix = prefix
	opts.indent = indent
}

func (r *Engine) SetSsl(certPath, keyPath, CAPath string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		zlog.Error("load keys fail", err)
		return nil, err
	}

	caData, err := ioutil.ReadFile(CAPath)
	if err != nil {
		zlog.Error("read ca fail", err)
		return nil, err
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	trans := r.getTransport()
	if trans == nil {
		return nil, ErrTransEmpty
	}

	trans.TLSClientConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
	}
	return trans.TLSClientConfig, nil
}

func (r *Engine) getTransport() *http.Transport {
	trans, _ := r.Client().Transport.(*http.Transport)
	return trans
}

func (r *Engine) getXMLEncOpts() *xmlEncOpts {
	if r.xmlEncOpts == nil {
		r.xmlEncOpts = &xmlEncOpts{}
	}
	return r.xmlEncOpts
}
