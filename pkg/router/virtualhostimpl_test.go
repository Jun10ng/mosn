/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"strings"
	"testing"

	"github.com/alipay/sofa-mosn/pkg/api/v2"
	"github.com/alipay/sofa-mosn/pkg/protocol"
)

// Prefix > Path > Regex
func TestRouterPriority(t *testing.T) {
	// only prefix is valid
	prefixRouter := v2.Router{}
	prefixRouter.Match = v2.RouterMatch{
		Prefix: "/foo",
		Path:   "/foo.html",
		Regex:  ".*",
	}
	prefixRouter.Route = v2.RouteAction{RouterActionConfig: v2.RouterActionConfig{ClusterName: "test"}}
	prefixVitrualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{prefixRouter},
	}, false)
	if len(prefixVitrualHost.routes) != 1 {
		t.Errorf("routes should have only one, but got :%v\n", len(prefixVitrualHost.routes))
	}
	if _, ok := prefixVitrualHost.routes[0].(*PrefixRouteRuleImpl); !ok {
		t.Error("cannot get a prefix route rule")
	}
	// only path is valid
	pathRouter := v2.Router{}
	pathRouter.Match = v2.RouterMatch{
		Path:  "/foo.html",
		Regex: ".*",
	}
	pathRouter.Route = v2.RouteAction{RouterActionConfig: v2.RouterActionConfig{ClusterName: "test"}}
	pathVirtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{pathRouter},
	}, false)
	if len(pathVirtualHost.routes) != 1 {
		t.Errorf("routes should have only one, but got :%v\n", len(pathVirtualHost.routes))
	}
	if _, ok := pathVirtualHost.routes[0].(*PathRouteRuleImpl); !ok {
		t.Error("cannot get a path route rule")
	}
}

// the first matched route will be used
func TestRouterOrder(t *testing.T) {
	prefixrouter := v2.Router{}
	prefixrouter.Match = v2.RouterMatch{
		Prefix: "/foo",
	}
	prefixrouter.Route = v2.RouteAction{
		RouterActionConfig: v2.RouterActionConfig{
			ClusterName: "prefix",
		},
	}
	pathrouter := v2.Router{}
	pathrouter.Match = v2.RouterMatch{
		Path: "/foo1",
	}
	pathrouter.Route = v2.RouteAction{
		RouterActionConfig: v2.RouterActionConfig{
			ClusterName: "path",
		},
	}
	regrouter := v2.Router{}
	regrouter.Match = v2.RouterMatch{
		Regex: "/foo[0-9]+",
	}
	regrouter.Route = v2.RouteAction{
		RouterActionConfig: v2.RouterActionConfig{
			ClusterName: "regexp",
		},
	}
	// path "/foo1" match all of the router, the path router should be matched
	// path "/foo11" match prefix and regexp router, the regexp router should be matched
	// path "/foo" match prefix router only
	routers := []v2.Router{pathrouter, regrouter, prefixrouter}
	testCases := []struct {
		path        string
		clustername string
	}{
		{"/foo1", "path"},
		{"/foo11", "regexp"},
		{"/foo", "prefix"},
	}
	virtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: routers,
	}, false)
	for i, tc := range testCases {
		headers := protocol.CommonHeader(map[string]string{
			strings.ToLower(protocol.MosnHeaderPathKey): tc.path,
		})
		rt := virtualHost.GetRouteFromEntries(headers, 1)
		if rt == nil || rt.RouteRule().ClusterName() != tc.clustername {
			t.Errorf("#%d route unexpected result\n", i)
		}
	}
	//prefix router first, only prefix will be matched
	prefixVirtualHost, _ := NewVirtualHostImpl(&v2.VirtualHost{
		Name:    "test",
		Domains: []string{"*"},
		Routers: []v2.Router{prefixrouter, regrouter, pathrouter},
	}, false)
	for i, tc := range testCases {
		headers := protocol.CommonHeader(map[string]string{
			strings.ToLower(protocol.MosnHeaderPathKey): tc.path,
		})
		rt := prefixVirtualHost.GetRouteFromEntries(headers, 1)
		if rt == nil || rt.RouteRule().ClusterName() != "prefix" {
			t.Errorf("#%d route unexpected result\n", i)
		}
	}

}
