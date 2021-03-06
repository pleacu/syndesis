/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package io.syndesis.integration.runtime.sb.tracing;

import static org.assertj.core.api.Assertions.assertThat;
import org.apache.camel.ExtendedCamelContext;
import org.apache.camel.spring.boot.CamelAutoConfiguration;
import org.apache.camel.spring.boot.CamelContextConfiguration;
import org.junit.Test;
import org.junit.runner.RunWith;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.context.ApplicationContext;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.context.junit4.SpringRunner;
import io.opentracing.Tracer;
import io.syndesis.common.jaeger.JaegerConfiguration;
import io.syndesis.integration.runtime.sb.IntegrationRuntimeAutoConfiguration;
import io.syndesis.integration.runtime.tracing.TracingLogListener;

@DirtiesContext
@RunWith(SpringRunner.class)
@SpringBootTest(
    classes = {
        CamelAutoConfiguration.class,
        TracingAutoConfiguration.class,
        IntegrationRuntimeAutoConfiguration.class,
        JaegerConfiguration.class,
    },
    properties = {
        "spring.main.banner-mode = off",
        "logging.level.io.syndesis.integration.runtime = DEBUG",
        "jaeger.service.name = IntegrationTracingEnabledTest",
        "syndesis.integration.runtime.tracing.enabled = true",
    }
)
public class IntegrationTracingEnabledTest {
    @Autowired
    private ApplicationContext applicationContext;
    @Autowired
    private ExtendedCamelContext camelContext;

    @Test
    public void testContextConfiguration() {
        assertThat(applicationContext.getBeansOfType(CamelContextConfiguration.class).values())
            .hasAtLeastOneElementOfType(TracingCamelContextConfiguration.class);
        assertThat(applicationContext.getBeansOfType(CamelContextConfiguration.class)).hasSize(2);
        assertThat(applicationContext.getBeansOfType(Tracer.class)).hasSize(1);
        assertThat(camelContext.getLogListeners()).hasAtLeastOneElementOfType(TracingLogListener.class);
    }
}
