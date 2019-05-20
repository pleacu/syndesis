import { Alert, Button, Col, Row } from 'patternfly-react';
import * as React from 'react';

export interface IOAuthAppExpanderBodyProps {
  showSuccess: boolean;
  disableSave: boolean;
  disableRemove: boolean;
  onSave: () => void;
  onRemove: () => void;
  children: React.ReactNode;
  i18nRemoveButtonText: string;
  i18nSaveButtonText: string;
  i18nAlertTitle: string;
  i18nAlertDetail: string;
}

export class OAuthAppExpanderBody extends React.Component<
  IOAuthAppExpanderBodyProps
> {
  constructor(props: IOAuthAppExpanderBodyProps) {
    super(props);
  }
  public render() {
    return (
      <>
        {this.props.showSuccess && (
          <Row>
            <Col xs={11}>
              <Alert type={'success'}>
                <strong>{this.props.i18nAlertTitle}</strong>&nbsp;
                {this.props.i18nAlertDetail}
              </Alert>
            </Col>
          </Row>
        )}
        <Row>
          <Col xs={6} xsOffset={5}>
            {this.props.children}
          </Col>
        </Row>
        <Row>
          <Col xs={6} xsOffset={5}>
            <>
              <Button
                data-testid={'oauth-app-expander-body-save'}
                bsStyle="primary"
                onClick={this.props.onSave}
                disabled={this.props.disableSave}
              >
                {this.props.i18nSaveButtonText}
              </Button>{' '}
              <Button
                data-testid={'oauth-app-expander-body-remove'}
                onClick={this.props.onRemove}
                disabled={this.props.disableRemove}
              >
                {this.props.i18nRemoveButtonText}
              </Button>
            </>
          </Col>
        </Row>
      </>
    );
  }
}
