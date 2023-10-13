import * as React from 'react';
import {injectable, postConstruct, inject} from '@theia/core/shared/inversify';
import {PreferenceService} from '@theia/core/lib/browser';
import {GettingStartedWidget} from "@theia/getting-started/lib/browser/getting-started-widget";
import {WindowService} from "@theia/core/lib/browser/window/window-service";
import {LogoIcon, QuickStart, Preference} from "./components";

@injectable()
export class CodigoCidlQuickstartWidget extends GettingStartedWidget {

    @inject(WindowService)
    protected readonly windowService: WindowService;

    @inject(PreferenceService)
    protected readonly preferenceService: PreferenceService;

    @postConstruct()
    protected async init(): Promise<void> {
        await super.init();
        await this.preferenceService.ready;
        this.update();
    }

    protected render(): React.ReactNode {
        return <div
            className='flex flex-row p-4 space-x-4 w-full h-full text-white border-l border-r border-lightGray justify-between bg-deepDarkGray overflow-hidden text-sm'>
            <div className="w-4/12 flex flex-col space-y-4">
                <div className={"flex flex-col items-start space-y-6"}>
                    <LogoIcon/>

                    <div className={"flex flex-col items-start space-y-3 text-gray-300 tracking-wide"}>
                        <p>
                            Welcome to Código Studio, our web-based development environment. Código Studio provides
                            comprehensive tools and programs for developers to create projects using our Código
                            Interface Description Language.
                        </p>

                        <p>
                            The starting point for blockchain developers using the Código platform is the Código
                            Interface Description Language (CIDL). The CIDL is a formal language used to describe the
                            interface of a blockchain smart contract.
                        </p>
                    </div>

                    <div className={"flex flex-col text-gray-300"}>
                        <h3 className={"font-bold text-lg text-gray-400"}>Useful links</h3>
                        <ul className={"space-y-0.5"}>
                            <li><a href={"https://codigo.ai"} target={"_blank"}>Home</a></li>
                            <li><a href={"https://docs.codigo.ai"} target={"_blank"}>Documentation</a></li>
                        </ul>
                    </div>
                </div>
            </div>

            <div className="w-8/12 border-l border-lightGray overflow-y-auto pl-4 space-y-4">
                <QuickStart/>
                <Preference preferenceService={this.preferenceService}/>
            </div>
        </div>;
    }
}



