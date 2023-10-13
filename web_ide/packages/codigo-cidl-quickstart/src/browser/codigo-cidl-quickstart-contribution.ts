import {inject, injectable} from '@theia/core/shared/inversify';
import {CodigoCidlQuickstartWidget} from './codigo-cidl-quickstart-widget';
import {
    AbstractViewContribution,
    FrontendApplication,
    FrontendApplicationContribution,
    PreferenceService
} from '@theia/core/lib/browser';
import {Command, CommandRegistry} from '@theia/core/lib/common/command';
import { GettingStartedWidget } from '@theia/getting-started/lib/browser/getting-started-widget';
import {FrontendApplicationStateService} from "@theia/core/lib/browser/frontend-application-state";
import {CodigoPreferences} from "./codigo-cidl-quickstart-preferences";
import {MenuModelRegistry} from "@theia/core";
import {MenuPath} from "@theia/core/lib/common/menu";
import {CommonMenus} from "@theia/core/lib/browser/common-frontend-contribution";
import {WindowService} from "@theia/core/lib/browser/window/window-service";

export namespace CodigoMenus {
    export const CODIGO_HELP: MenuPath = [...CommonMenus.HELP, 'codigo'];
}

export namespace CodigoCommands {
    export const CATEGORY = 'Codigo';

    export const GET_STARTED: Command = {
        id: 'codigo:get_started',
        category: CATEGORY,
        label: 'Get Started'
    };
    
    export const DOCUMENTATION: Command = {
        id: 'codigo:documentation',
        category: CATEGORY,
        label: 'Documentation'
    };

    export const TERMS: Command = {
        id: 'codigo:tos',
        category: CATEGORY,
        label: 'Terms of Service'
    };

    export const PRIVACY: Command = {
        id: 'codigo:privacy',
        category: CATEGORY,
        label: 'Privacy Policy'
    };
}

@injectable()
export class CodigoCidlQuickstartContribution extends AbstractViewContribution<CodigoCidlQuickstartWidget> implements FrontendApplicationContribution {

    static DOCUMENTATION_URL = 'https://docs.codigo.ai';
    static TERMS_URL = 'https://docs.codigo.ai/terms';
    static PRIVACY_URL = 'https://docs.codigo.ai/privacy';

    @inject(WindowService)
    protected readonly windowService: WindowService;
    
    @inject(FrontendApplicationStateService)
    protected readonly stateService: FrontendApplicationStateService;

    @inject(PreferenceService)
    protected readonly preferenceService: PreferenceService;
    
    constructor() {
        super({
            widgetId: GettingStartedWidget.ID,
            widgetName: GettingStartedWidget.LABEL,
            defaultWidgetOptions: {area: 'main'}
        });
    }

    async onStart(app: FrontendApplication): Promise<void> {
        this.stateService.reachedState('ready').then(
            () => this.preferenceService.ready.then(() => {
                const showQuickStartPage: boolean = this.preferenceService.get(CodigoPreferences.alwaysShowQuickStartPage, true);

                if (showQuickStartPage) {
                    this.openView({reveal: true, activate: true});
                }
            })
        );
    }

    registerCommands(commands: CommandRegistry): void {
        commands.registerCommand(CodigoCommands.GET_STARTED, {
            execute: () => this.openView({reveal: true, activate: true})
        });
        commands.registerCommand(CodigoCommands.DOCUMENTATION, {
            execute: () => this.windowService.openNewWindow(CodigoCidlQuickstartContribution.DOCUMENTATION_URL, { external: true })
        });
        commands.registerCommand(CodigoCommands.TERMS, {
            execute: () => this.windowService.openNewWindow(CodigoCidlQuickstartContribution.TERMS_URL, { external: true })
        });
        commands.registerCommand(CodigoCommands.PRIVACY, {
            execute: () => this.windowService.openNewWindow(CodigoCidlQuickstartContribution.PRIVACY_URL, { external: true })
        });
    }

    registerMenus(menus: MenuModelRegistry): void {
        menus.registerMenuAction(CodigoMenus.CODIGO_HELP, {
            commandId: CodigoCommands.GET_STARTED.id,
            label: CodigoCommands.GET_STARTED.label,
            order: '1'
        });
        menus.registerMenuAction(CodigoMenus.CODIGO_HELP, {
            commandId: CodigoCommands.DOCUMENTATION.id,
            label: CodigoCommands.DOCUMENTATION.label,
            order: '2'
        });
        menus.registerMenuAction(CodigoMenus.CODIGO_HELP, {
            commandId: CodigoCommands.TERMS.id,
            label: CodigoCommands.TERMS.label,
            order: '3'
        });
        menus.registerMenuAction(CodigoMenus.CODIGO_HELP, {
            commandId: CodigoCommands.PRIVACY.id,
            label: CodigoCommands.PRIVACY.label,
            order: '4'
        });
    }
}
