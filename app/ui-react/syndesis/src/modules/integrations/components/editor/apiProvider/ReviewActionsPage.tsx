import {
  useApiProviderIntegration,
  useApiProviderSummary,
} from '@syndesis/api';
import * as H from '@syndesis/history';
import { Integration } from '@syndesis/models';
import {
  ButtonLink,
  IntegrationEditorLayout,
  OpenApiReviewActions,
  PageLoader,
  PageSection,
} from '@syndesis/ui';
import { useRouteData, WithLoader } from '@syndesis/utils';
import * as React from 'react';
import { Translation } from 'react-i18next';
import { ApiError, PageTitle } from '../../../../../shared';
import {
  IApiProviderReviewActionsRouteState,
  IBaseApiProviderRouteParams,
  IPageWithEditorBreadcrumb,
} from '../interfaces';

export interface IReviewActionsPageProps extends IPageWithEditorBreadcrumb {
  cancelHref: (
    p: IBaseApiProviderRouteParams,
    s: IApiProviderReviewActionsRouteState
  ) => H.LocationDescriptor;
  editHref: (
    p: IBaseApiProviderRouteParams,
    s: IApiProviderReviewActionsRouteState
  ) => H.LocationDescriptor;
  nextHref: (
    integration: Integration,
    p: IBaseApiProviderRouteParams,
    s: IApiProviderReviewActionsRouteState
  ) => H.LocationDescriptorObject;
}

/**
 * This is the page where a user reviews the actions that have been
 * extracted from the API specification previously created or provided
 * earlier in the API Provider editor.
 */
export const ReviewActionsPage: React.FunctionComponent<
  IReviewActionsPageProps
> = ({ cancelHref, editHref, nextHref, getBreadcrumb }) => {
  const [nextDisabled, setNextDisabled] = React.useState(false);
  const { params, state, history } = useRouteData<
    IBaseApiProviderRouteParams,
    IApiProviderReviewActionsRouteState
  >();
  const { apiSummary, loading, error } = useApiProviderSummary(
    state.specification
  );
  const getIntegration = useApiProviderIntegration();

  const onNext = async () => {
    setNextDisabled(true);
    try {
      const integration = await getIntegration(
        apiSummary!.configuredProperties!.specification
      );
      integration.id = state.integration.id;
      integration.name = '';
      history.push(nextHref(integration, params, state));
    } catch (e) {
      // todo show the error?
    }
    setNextDisabled(false);
  };

  return (
    <Translation ns={['integrations', 'shared']}>
      {t => (
        <>
          <PageTitle
            title={t('integrations:apiProvider:reviewActions:title')}
          />
          <IntegrationEditorLayout
            title={t('integrations:apiProvider:reviewActions:title')}
            description={t(
              'integrations:apiProvider:reviewActions:description'
            )}
            toolbar={getBreadcrumb(
              t('integrations:apiProvider:reviewActions:title'),
              params,
              state
            )}
            content={
              <PageSection>
                <WithLoader
                  loading={loading}
                  loaderChildren={<PageLoader />}
                  error={error !== false}
                  errorChildren={<ApiError />}
                >
                  {() => (
                    <>
                      <OpenApiReviewActions
                        i18nApiDefinitionHeading={'API DEFINITION'}
                        i18nDescriptionLabel={'Description'}
                        i18nImportedHeading={'IMPORTED'}
                        i18nNameLabel={'Name'}
                        apiProviderDescription={apiSummary!.description}
                        apiProviderName={apiSummary!.name}
                        i18nOperationsHtmlMessage={`<strong>${
                          apiSummary!.actionsSummary!.totalActions
                        }</strong> operations`}
                        i18nWarningsHeading={
                          apiSummary!.warnings
                            ? `WARNINGS <strong> ${
                                apiSummary!.warnings.length
                              }</strong>`
                            : undefined
                        }
                        warningMessages={
                          apiSummary!.warnings
                            ? apiSummary!.warnings.map(
                                warning => (warning as any).message
                              )
                            : undefined
                        }
                      />
                      <div>
                        <ButtonLink
                          href={editHref(params, {
                            ...state,
                            specification: apiSummary!.configuredProperties!
                              .specification,
                          })}
                        >
                          Review/Edit
                        </ButtonLink>
                        <ButtonLink
                          onClick={onNext}
                          disabled={nextDisabled}
                          as={'primary'}
                        >
                          Next
                        </ButtonLink>
                      </div>
                    </>
                  )}
                </WithLoader>
              </PageSection>
            }
            cancelHref={cancelHref(params, state)}
          />
        </>
      )}
    </Translation>
  );
};