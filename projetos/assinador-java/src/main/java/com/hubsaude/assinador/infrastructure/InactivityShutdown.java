package com.hubsaude.assinador.infrastructure;

import org.springframework.boot.ApplicationArguments;
import org.springframework.boot.ApplicationRunner;
import org.springframework.stereotype.Component;

/**
 * Encerra o servidor automaticamente após um período de inatividade configurável.
 *
 * <p>Ativado pela variável de ambiente {@code HUBSAUDE_TIMEOUT_MINUTES}. Quando
 * ausente ou zero, nenhum watchdog é iniciado. A thread de monitoramento verifica
 * a cada 30 segundos se o tempo sem requisições superou o limite; se sim, chama
 * {@code System.exit(0)}.
 *
 * <p>O CLI passa esta variável ao iniciar o servidor com {@code assinatura start --timeout N}.
 */
@Component
public class InactivityShutdown implements ApplicationRunner {

    private final RequestTimestamp requestTimestamp;

    public InactivityShutdown(RequestTimestamp requestTimestamp) {
        this.requestTimestamp = requestTimestamp;
    }

    @Override
    public void run(ApplicationArguments args) {
        String env = System.getenv("HUBSAUDE_TIMEOUT_MINUTES");
        if (env == null || env.isBlank()) return;

        int minutes;
        try {
            minutes = Integer.parseInt(env.trim());
            if (minutes <= 0) return;
        } catch (NumberFormatException e) {
            return;
        }

        long timeoutMs = (long) minutes * 60_000L;
        Thread watchdog = new Thread(() -> {
            while (!Thread.currentThread().isInterrupted()) {
                try {
                    Thread.sleep(30_000);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    return;
                }
                if (System.currentTimeMillis() - requestTimestamp.get() > timeoutMs) {
                    System.err.printf("Servidor encerrando por inatividade (timeout=%dmin)%n", minutes);
                    System.exit(0);
                }
            }
        });
        watchdog.setDaemon(true);
        watchdog.setName("inactivity-watchdog");
        watchdog.start();
    }
}
