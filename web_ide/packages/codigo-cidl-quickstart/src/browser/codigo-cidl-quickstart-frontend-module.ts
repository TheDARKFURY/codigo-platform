import { ContainerModule } from '@theia/core/shared/inversify';
import { CodigoCidlQuickstartWidget } from './codigo-cidl-quickstart-widget';
import { CodigoCidlQuickstartContribution } from './codigo-cidl-quickstart-contribution';
import {
    bindViewContribution,
    FrontendApplicationContribution,
    PreferenceContribution,
    WidgetFactory
} from '@theia/core/lib/browser';

import '../../src/browser/style/index.css';
import {codigoPreferenceSchema} from "./codigo-cidl-quickstart-preferences";
import {GettingStartedWidget} from "@theia/getting-started/lib/browser/getting-started-widget";

import '../../src/browser/style/bundle.css';
import '../../src/browser/style/global.css';

export default new ContainerModule(bind => {
    bindViewContribution(bind, CodigoCidlQuickstartContribution);
    bind(FrontendApplicationContribution).toService(CodigoCidlQuickstartContribution);
    bind(CodigoCidlQuickstartWidget).toSelf();
    bind(WidgetFactory).toDynamicValue(ctx => ({
        id: GettingStartedWidget.ID,
        createWidget: () => ctx.container.get<CodigoCidlQuickstartWidget>(CodigoCidlQuickstartWidget)
    })).inSingletonScope();
    bind(PreferenceContribution).toConstantValue({ schema: codigoPreferenceSchema });
});
