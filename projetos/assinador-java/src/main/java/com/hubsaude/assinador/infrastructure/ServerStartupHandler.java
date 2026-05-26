package com.hubsaude.assinador.infrastructure;

import org.springframework.boot.web.context.WebServerInitializedEvent;
import org.springframework.context.ApplicationListener;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

@Component
public class ServerStartupHandler implements ApplicationListener<WebServerInitializedEvent> {

    @Override
    public void onApplicationEvent(WebServerInitializedEvent event) {
        int port = event.getWebServer().getPort();
        long pid = ProcessHandle.current().pid();

        Path hubSaudeDir = Path.of(System.getProperty("user.home"), ".hubsaude");
        try {
            Files.createDirectories(hubSaudeDir);
            String content = String.format("{\"pid\":%d,\"port\":%d}%n", pid, port);
            Files.writeString(hubSaudeDir.resolve("assinador.pid"), content);
        } catch (IOException e) {
            System.err.println("Aviso: não foi possível registrar PID/porta: " + e.getMessage());
        }

        System.err.printf("Servidor iniciado na porta %d (PID %d)%n", port, pid);
    }
}
