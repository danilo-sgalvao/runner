package com.hubsaude.assinador;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import com.hubsaude.assinador.presentation.cli.CliPresenter;
import com.hubsaude.assinador.presentation.cli.CliRunner;
import org.springframework.boot.SpringApplication;

import java.util.Properties;

public class AssinadorApplication {

    public static void main(String[] args) {
        if (args.length > 0 && "serve".equals(args[0])) {
            int port = 8080;
            for (int i = 1; i < args.length; i++) {
                if ("--port".equals(args[i]) && i + 1 < args.length) {
                    try {
                        port = Integer.parseInt(args[++i]);
                    } catch (NumberFormatException e) {
                        System.err.println("Erro: porta inválida: " + args[i]);
                        System.exit(1);
                    }
                }
            }

            SpringApplication app = new SpringApplication(WebApplication.class);
            Properties props = new Properties();
            props.setProperty("server.port", String.valueOf(port));
            app.setDefaultProperties(props);
            app.run(new String[0]);
            return;
        }

        SignatureService service         = new FakeSignatureService();
        RequestValidator validator       = new RequestValidator();
        SignUseCase      signUseCase     = new SignUseCase(service, validator);
        ValidateUseCase  validateUseCase = new ValidateUseCase(service, validator);
        CliPresenter     presenter       = new CliPresenter();
        CliRunner        runner          = new CliRunner(signUseCase, validateUseCase, presenter);

        runner.run(args);
    }
}
