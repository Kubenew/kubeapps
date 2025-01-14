/*
Copyright © 2021 VMware
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/packages/v1alpha1"
	plugins "github.com/kubeapps/kubeapps/cmd/kubeapps-apis/gen/core/plugins/v1alpha1"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/plugin_test"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	globalPackagingNamespace = "kubeapps"
)

var mockedPackagingPlugin1 = makeDefaultTestPackagingPlugin("mock1")
var mockedPackagingPlugin2 = makeDefaultTestPackagingPlugin("mock2")
var mockedNotFoundPackagingPlugin = makeOnlyStatusTestPackagingPlugin("bad-plugin", codes.NotFound)

var ignoreUnexportedOpts = cmpopts.IgnoreUnexported(
	corev1.AvailablePackageDetail{},
	corev1.AvailablePackageReference{},
	corev1.AvailablePackageSummary{},
	corev1.Context{},
	corev1.GetAvailablePackageDetailResponse{},
	corev1.GetAvailablePackageSummariesResponse{},
	corev1.GetAvailablePackageVersionsResponse{},
	corev1.GetInstalledPackageDetailResponse{},
	corev1.GetInstalledPackageSummariesResponse{},
	corev1.CreateInstalledPackageResponse{},
	corev1.UpdateInstalledPackageResponse{},
	corev1.InstalledPackageDetail{},
	corev1.InstalledPackageReference{},
	corev1.InstalledPackageStatus{},
	corev1.InstalledPackageSummary{},
	corev1.Maintainer{},
	corev1.PackageAppVersion{},
	corev1.VersionReference{},
	plugins.Plugin{},
)

func makeDefaultTestPackagingPlugin(pluginName string) *pkgsPluginWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugin_test.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.AvailablePackageSummaries = []*corev1.AvailablePackageSummary{
		plugin_test.MakeAvailablePackageSummary("pkg-2", pluginDetails),
		plugin_test.MakeAvailablePackageSummary("pkg-1", pluginDetails),
	}
	packagingPluginServer.AvailablePackageDetail = plugin_test.MakeAvailablePackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.InstalledPackageSummaries = []*corev1.InstalledPackageSummary{
		plugin_test.MakeInstalledPackageSummary("pkg-2", pluginDetails),
		plugin_test.MakeInstalledPackageSummary("pkg-1", pluginDetails),
	}
	packagingPluginServer.InstalledPackageDetail = plugin_test.MakeInstalledPackageDetail("pkg-1", pluginDetails)
	packagingPluginServer.PackageAppVersions = []*corev1.PackageAppVersion{
		plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgUpdateVersion),
		plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgVersion),
	}
	packagingPluginServer.NextPageToken = "1"
	packagingPluginServer.Categories = []string{plugin_test.DefaultCategory}

	return &pkgsPluginWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func makeOnlyStatusTestPackagingPlugin(pluginName string, statusCode codes.Code) *pkgsPluginWithServer {
	pluginDetails := &plugins.Plugin{Name: pluginName, Version: "v1alpha1"}
	packagingPluginServer := &plugin_test.TestPackagingPluginServer{Plugin: pluginDetails}

	packagingPluginServer.Status = statusCode

	return &pkgsPluginWithServer{
		plugin: pluginDetails,
		server: packagingPluginServer,
	}
}

func TestGetAvailablePackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageSummariesRequest
		expectedResponse  *corev1.GetAvailablePackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories: []string{"cat-1"},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (first page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "0", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (proper PageSize) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "0", PageSize: 4},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "1",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (last page - 1) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "3", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "4",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (last page) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "3", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{
					plugin_test.MakeAvailablePackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
				Categories:    []string{"cat-1"},
				NextPageToken: "4",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should successfully call and paginate (last page + 1) the core GetAvailablePackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
				PaginationOptions: &corev1.PaginationOptions{PageToken: "4", PageSize: 1},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				Categories:                []string{"cat-1"},
				NextPageToken:             "",
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageSummaries operation when the package is not present in a plugin",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageSummariesResponse{
				AvailablePackageSummaries: []*corev1.AvailablePackageSummary{},
				Categories:                []string{""},
			},
			statusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			availablePackageSummaries, err := server.GetAvailablePackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := availablePackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetAvailablePackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageDetailRequest
		expectedResponse  *corev1.GetAvailablePackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageDetail operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &corev1.GetAvailablePackageDetailResponse{
				AvailablePackageDetail: plugin_test.MakeAvailablePackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageDetailRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
				PkgVersion: "",
			},

			expectedResponse: &corev1.GetAvailablePackageDetailResponse{},
			statusCode:       codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			availablePackageDetail, err := server.GetAvailablePackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := availablePackageDetail, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetInstalledPackageSummaries(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageSummariesRequest
		expectedResponse  *corev1.GetInstalledPackageSummariesResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageSummaries operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{
					plugin_test.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin1.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-1", mockedPackagingPlugin2.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin1.plugin),
					plugin_test.MakeInstalledPackageSummary("pkg-2", mockedPackagingPlugin2.plugin),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageSummaries operation when the package is not present in a plugin",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetInstalledPackageSummariesRequest{
				Context: &corev1.Context{
					Cluster:   "",
					Namespace: globalPackagingNamespace,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageSummariesResponse{
				InstalledPackageSummaries: []*corev1.InstalledPackageSummary{},
			},
			statusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			installedPackageSummaries, err := server.GetInstalledPackageSummaries(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPackageSummaries, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetInstalledPackageDetail(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetInstalledPackageDetailRequest
		expectedResponse  *corev1.GetInstalledPackageDetailResponse
	}{
		{
			name: "it should successfully call the core GetInstalledPackageDetail operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{
				InstalledPackageDetail: plugin_test.MakeInstalledPackageDetail("pkg-1", mockedPackagingPlugin1.plugin),
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetInstalledPackageDetail operation when the package is not present in a plugin",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetInstalledPackageDetailRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "pkg-1",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},

			expectedResponse: &corev1.GetInstalledPackageDetailResponse{},
			statusCode:       codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			installedPackageDetail, err := server.GetInstalledPackageDetail(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPackageDetail, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestGetAvailablePackageVersions(t *testing.T) {
	testCases := []struct {
		name              string
		configuredPlugins []*pkgsPluginWithServer
		statusCode        codes.Code
		request           *corev1.GetAvailablePackageVersionsRequest
		expectedResponse  *corev1.GetAvailablePackageVersionsResponse
	}{
		{
			name: "it should successfully call the core GetAvailablePackageVersions operation",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedPackagingPlugin2,
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedPackagingPlugin1.plugin,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{
					plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgUpdateVersion),
					plugin_test.MakePackageAppVersion(plugin_test.DefaultAppVersion, plugin_test.DefaultPkgVersion),
				},
			},
			statusCode: codes.OK,
		},
		{
			name: "it should fail when calling the core GetAvailablePackageVersions operation when the package is not present in a plugin",
			configuredPlugins: []*pkgsPluginWithServer{
				mockedPackagingPlugin1,
				mockedNotFoundPackagingPlugin,
			},
			request: &corev1.GetAvailablePackageVersionsRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Context: &corev1.Context{
						Cluster:   "",
						Namespace: globalPackagingNamespace,
					},
					Identifier: "test",
					Plugin:     mockedNotFoundPackagingPlugin.plugin,
				},
			},

			expectedResponse: &corev1.GetAvailablePackageVersionsResponse{
				PackageAppVersions: []*corev1.PackageAppVersion{},
			},
			statusCode: codes.NotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := &packagesServer{
				plugins: tc.configuredPlugins,
			}
			AvailablePackageVersions, err := server.GetAvailablePackageVersions(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := AvailablePackageVersions, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestCreateInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.CreateInstalledPackageRequest
		expectedResponse  *corev1.CreateInstalledPackageResponse
	}{
		{
			name: "installs the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
			expectedResponse: &corev1.CreateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.CreateInstalledPackageRequest{
				AvailablePackageRef: &corev1.AvailablePackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
				TargetContext: &corev1.Context{
					Cluster:   "default",
					Namespace: "my-ns",
				},
				Name: "installed-pkg-1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []*pkgsPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, &pkgsPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				plugins: configuredPluginServers,
			}

			installedPkgResponse, err := server.CreateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := installedPkgResponse, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestUpdateInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.UpdateInstalledPackageRequest
		expectedResponse  *corev1.UpdateInstalledPackageResponse
	}{
		{
			name: "updates the package using the correct plugin",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
			expectedResponse: &corev1.UpdateInstalledPackageResponse{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.UpdateInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []*pkgsPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, &pkgsPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				plugins: configuredPluginServers,
			}

			updatedPkgResponse, err := server.UpdateInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}

			if tc.statusCode == codes.OK {
				if got, want := updatedPkgResponse, tc.expectedResponse; !cmp.Equal(got, want, ignoreUnexportedOpts) {
					t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got, ignoreUnexportedOpts))
				}
			}
		})
	}
}

func TestDeleteInstalledPackage(t *testing.T) {

	testCases := []struct {
		name              string
		configuredPlugins []*plugins.Plugin
		statusCode        codes.Code
		request           *corev1.DeleteInstalledPackageRequest
	}{
		{
			name: "deletes the package",
			configuredPlugins: []*plugins.Plugin{
				{Name: "plugin-1", Version: "v1alpha1"},
				{Name: "plugin-1", Version: "v1alpha2"},
			},
			statusCode: codes.OK,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Context:    &corev1.Context{Cluster: "default", Namespace: "my-ns"},
					Identifier: "installed-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
		{
			name:       "returns invalid argument if plugin not specified in request",
			statusCode: codes.InvalidArgument,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
				},
			},
		},
		{
			name:       "returns internal error if unable to find the plugin",
			statusCode: codes.Internal,
			request: &corev1.DeleteInstalledPackageRequest{
				InstalledPackageRef: &corev1.InstalledPackageReference{
					Identifier: "available-pkg-1",
					Plugin:     &plugins.Plugin{Name: "plugin-1", Version: "v1alpha1"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configuredPluginServers := []*pkgsPluginWithServer{}
			for _, p := range tc.configuredPlugins {
				configuredPluginServers = append(configuredPluginServers, &pkgsPluginWithServer{
					plugin: p,
					server: plugin_test.TestPackagingPluginServer{Plugin: p},
				})
			}

			server := &packagesServer{
				plugins: configuredPluginServers,
			}

			_, err := server.DeleteInstalledPackage(context.Background(), tc.request)

			if got, want := status.Code(err), tc.statusCode; got != want {
				t.Fatalf("got: %+v, want: %+v, err: %+v", got, want, err)
			}
		})
	}
}
