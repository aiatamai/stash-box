import type { FC } from "react";
import { useState } from "react";
import { Button, Form, Row, Col, Alert } from "react-bootstrap";
import { useForm, Controller } from "react-hook-form";
import cx from "classnames";

import { useConfig, useUpdateConfig } from "src/graphql";

type ConfigFormData = {
  title: string;
  host_url: string;
  require_invite: boolean;
  require_activation: boolean;
  activation_expiry: number;
  email_cooldown: number;
  default_user_roles: string;
  vote_promotion_threshold: number;
  vote_application_threshold: number;
  voting_period: number;
  min_destructive_voting_period: number;
  vote_cron_interval: string;
  guidelines_url: string;
  edit_update_limit: number;
  require_scene_draft: boolean;
  require_tag_role: boolean;
  email_host: string;
  email_port: number;
  email_user: string;
  email_password: string;
  email_from: string;
  image_location: string;
  image_backend: string;
  image_jpeg_quality: number;
  image_max_size: number;
  image_resizing_enabled: boolean;
  image_resizing_cache_path: string;
  image_resizing_min_size: number;
  s3_endpoint: string;
  s3_bucket: string;
  s3_access_key: string;
  s3_secret: string;
  s3_max_dimension: number;
  postgres_max_open_conns: number;
  postgres_max_idle_conns: number;
  postgres_conn_max_lifetime: number;
  phash_distance: number;
  favicon_path: string;
  draft_time_limit: number;
  profiler_port: number;
  user_log_file: string;
  csp: string;
};

const Settings: FC = () => {
  const { loading, data } = useConfig();
  const [updateConfig] = useUpdateConfig();
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    control,
    formState: { errors, isSubmitting },
  } = useForm<ConfigFormData>({
    defaultValues: data?.getConfig
      ? {
          ...data.getConfig,
          default_user_roles: data.getConfig.default_user_roles.join(", "),
        }
      : undefined,
  });

  if (loading || !data) {
    return <div>Loading...</div>;
  }

  const onSubmit = async (formData: ConfigFormData) => {
    setSuccessMessage(null);
    setErrorMessage(null);

    try {
      const input = {
        ...formData,
        default_user_roles: formData.default_user_roles
          .split(",")
          .map((r) => r.trim())
          .filter((r) => r),
      };

      await updateConfig({ variables: { input } });
      setSuccessMessage(
        "Configuration updated successfully! Please restart the server for changes to take effect.",
      );
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : "Failed to update configuration",
      );
    }
  };

  return (
    <div>
      <h1>Settings</h1>
      <p className="lead">Configure your Stash-Box instance settings.</p>

      {successMessage && <Alert variant="success">{successMessage}</Alert>}
      {errorMessage && <Alert variant="danger">{errorMessage}</Alert>}

      <Form onSubmit={handleSubmit(onSubmit)}>
        {/* General Settings */}
        <h3 className="mt-4">General Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Title</Form.Label>
            <Form.Control
              className={cx({ "is-invalid": errors.title })}
              placeholder="Stash-Box"
              {...register("title")}
            />
            <Form.Text>Title of the instance, used in the page title.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Host URL</Form.Label>
            <Form.Control
              className={cx({ "is-invalid": errors.host_url })}
              placeholder="https://hostname.com"
              {...register("host_url")}
            />
            <Form.Text>
              Base URL for the server. Used when sending emails.
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Guidelines URL</Form.Label>
            <Form.Control
              placeholder="https://hostname.com/guidelines"
              {...register("guidelines_url")}
            />
            <Form.Text>
              URL to link to a set of guidelines for users contributing edits.
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Controller
              name="require_invite"
              control={control}
              render={({ field }) => (
                <Form.Check
                  type="checkbox"
                  label="Require Invite"
                  checked={field.value}
                  onChange={(e) => field.onChange(e.target.checked)}
                />
              )}
            />
            <Form.Text>
              If true, users are required to enter an invite key to create a new
              account.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Controller
              name="require_activation"
              control={control}
              render={({ field }) => (
                <Form.Check
                  type="checkbox"
                  label="Require Activation"
                  checked={field.value}
                  onChange={(e) => field.onChange(e.target.checked)}
                />
              )}
            />
            <Form.Text>
              If true, users are required to verify their email address before
              creating an account.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Controller
              name="require_scene_draft"
              control={control}
              render={({ field }) => (
                <Form.Check
                  type="checkbox"
                  label="Require Scene Draft"
                  checked={field.value}
                  onChange={(e) => field.onChange(e.target.checked)}
                />
              )}
            />
            <Form.Text>
              Whether to allow scene creation outside of draft submissions.
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Activation Expiry (seconds)</Form.Label>
            <Form.Control
              type="number"
              {...register("activation_expiry", { valueAsNumber: true })}
            />
            <Form.Text>
              Time after which an activation key expires. (Default: 7200 = 2 hours)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Email Cooldown (seconds)</Form.Label>
            <Form.Control
              type="number"
              {...register("email_cooldown", { valueAsNumber: true })}
            />
            <Form.Text>
              Time a user must wait before submitting another activation request.
              (Default: 300 = 5 minutes)
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Default User Roles</Form.Label>
            <Form.Control
              placeholder="READ, VOTE, EDIT"
              {...register("default_user_roles")}
            />
            <Form.Text>
              Comma-separated roles assigned to new users when registering.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Controller
              name="require_tag_role"
              control={control}
              render={({ field }) => (
                <Form.Check
                  type="checkbox"
                  label="Require Tag Role"
                  checked={field.value}
                  onChange={(e) => field.onChange(e.target.checked)}
                />
              )}
            />
            <Form.Text>Whether to require the EditTag role to edit tags.</Form.Text>
          </Form.Group>
        </Row>

        {/* Voting & Edit Settings */}
        <h3 className="mt-4">Voting & Edit Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Vote Promotion Threshold</Form.Label>
            <Form.Control
              type="number"
              {...register("vote_promotion_threshold", { valueAsNumber: true })}
            />
            <Form.Text>
              Number of approved edits before a user automatically gets VOTE role.
              Leave empty to disable.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Vote Application Threshold</Form.Label>
            <Form.Control
              type="number"
              {...register("vote_application_threshold", { valueAsNumber: true })}
            />
            <Form.Text>
              Number of same votes required for immediate application of an edit.
              (Default: 3)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Edit Update Limit</Form.Label>
            <Form.Control
              type="number"
              {...register("edit_update_limit", { valueAsNumber: true })}
            />
            <Form.Text>
              Number of times an edit can be updated by the creator. (Default: 1)
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Voting Period (seconds)</Form.Label>
            <Form.Control
              type="number"
              {...register("voting_period", { valueAsNumber: true })}
            />
            <Form.Text>
              Time before a voting period is closed. (Default: 345600 = 4 days)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Min Destructive Voting Period (seconds)</Form.Label>
            <Form.Control
              type="number"
              {...register("min_destructive_voting_period", {
                valueAsNumber: true,
              })}
            />
            <Form.Text>
              Minimum time that needs to pass before a destructive edit can be
              immediately applied. (Default: 172800 = 2 days)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Vote Cron Interval</Form.Label>
            <Form.Control
              placeholder="5m"
              {...register("vote_cron_interval")}
            />
            <Form.Text>
              Time between runs to close edits whose voting periods have ended.
              (Default: 5m)
            </Form.Text>
          </Form.Group>
        </Row>

        {/* Email Settings */}
        <h3 className="mt-4">Email Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Email Host</Form.Label>
            <Form.Control placeholder="smtp.example.com" {...register("email_host")} />
            <Form.Text>
              Address of the SMTP server. Required to send emails for activation
              and recovery.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Email Port</Form.Label>
            <Form.Control
              type="number"
              {...register("email_port", { valueAsNumber: true })}
            />
            <Form.Text>
              Port of the SMTP server. Only STARTTLS is supported. (Default: 25)
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Email User</Form.Label>
            <Form.Control
              placeholder="username"
              {...register("email_user")}
            />
            <Form.Text>Username for the SMTP server (optional).</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Email Password</Form.Label>
            <Form.Control
              type="text"
              placeholder="password"
              {...register("email_password")}
            />
            <Form.Text>Password for the SMTP server (optional).</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Email From</Form.Label>
            <Form.Control
              placeholder="noreply@example.com"
              {...register("email_from")}
            />
            <Form.Text>Email address from which to send emails.</Form.Text>
          </Form.Group>
        </Row>

        {/* Image Settings */}
        <h3 className="mt-4">Image Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Image Location</Form.Label>
            <Form.Control
              placeholder="/path/to/images"
              {...register("image_location")}
            />
            <Form.Text>Path to store images, for local image storage.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Image Backend</Form.Label>
            <Form.Select {...register("image_backend")}>
              <option value="file">File</option>
              <option value="s3">S3</option>
            </Form.Select>
            <Form.Text>Storage solution for images.</Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Image JPEG Quality</Form.Label>
            <Form.Control
              type="number"
              {...register("image_jpeg_quality", { valueAsNumber: true })}
            />
            <Form.Text>
              Quality setting when resizing JPEG images (0-100). (Default: 75)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Image Max Size</Form.Label>
            <Form.Control
              type="number"
              {...register("image_max_size", { valueAsNumber: true })}
            />
            <Form.Text>
              Max size of image if no size is specified. Omit to return full size.
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Controller
              name="image_resizing_enabled"
              control={control}
              render={({ field }) => (
                <Form.Check
                  type="checkbox"
                  label="Image Resizing Enabled"
                  checked={field.value}
                  onChange={(e) => field.onChange(e.target.checked)}
                />
              )}
            />
            <Form.Text>Whether to resize images shown in the frontend.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Image Resizing Cache Path</Form.Label>
            <Form.Control
              placeholder="/path/to/cache"
              {...register("image_resizing_cache_path")}
            />
            <Form.Text>
              Folder where resized images will be saved for later requests.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Image Resizing Min Size</Form.Label>
            <Form.Control
              type="number"
              {...register("image_resizing_min_size", { valueAsNumber: true })}
            />
            <Form.Text>Only resize images above a certain size.</Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Favicon Path</Form.Label>
            <Form.Control
              placeholder="/path/to/favicons"
              {...register("favicon_path")}
            />
            <Form.Text>
              Location where favicons for linked sites should be stored. Leave empty
              to disable.
            </Form.Text>
          </Form.Group>
        </Row>

        {/* S3 Settings */}
        <h3 className="mt-4">S3 Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>S3 Endpoint</Form.Label>
            <Form.Control
              placeholder="s3.amazonaws.com"
              {...register("s3_endpoint")}
            />
            <Form.Text>Hostname to S3 endpoint used for image storage.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>S3 Bucket</Form.Label>
            <Form.Control placeholder="my-bucket" {...register("s3_bucket")} />
            <Form.Text>Name of S3 bucket used to store images.</Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>S3 Access Key</Form.Label>
            <Form.Control type="text" {...register("s3_access_key")} />
            <Form.Text>Access key used for authentication.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>S3 Secret</Form.Label>
            <Form.Control type="text" {...register("s3_secret")} />
            <Form.Text>Secret access key used for authentication.</Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>S3 Max Dimension</Form.Label>
            <Form.Control
              type="number"
              {...register("s3_max_dimension", { valueAsNumber: true })}
            />
            <Form.Text>
              If set, a resized copy will be created for any image whose dimensions
              exceed this number.
            </Form.Text>
          </Form.Group>
        </Row>

        {/* Database Settings */}
        <h3 className="mt-4">Database Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Max Open Connections</Form.Label>
            <Form.Control
              type="number"
              {...register("postgres_max_open_conns", { valueAsNumber: true })}
            />
            <Form.Text>
              Maximum number of concurrent open connections to the database.
              (Default: 0 = unlimited)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Max Idle Connections</Form.Label>
            <Form.Control
              type="number"
              {...register("postgres_max_idle_conns", { valueAsNumber: true })}
            />
            <Form.Text>
              Maximum number of concurrent idle database connections. (Default: 0 =
              unlimited)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Connection Max Lifetime (minutes)</Form.Label>
            <Form.Control
              type="number"
              {...register("postgres_conn_max_lifetime", { valueAsNumber: true })}
            />
            <Form.Text>
              Maximum lifetime in minutes before a connection is released. (Default:
              0 = unlimited)
            </Form.Text>
          </Form.Group>
        </Row>

        {/* Other Settings */}
        <h3 className="mt-4">Other Settings</h3>
        <hr />

        <Row>
          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>pHash Distance</Form.Label>
            <Form.Control
              type="number"
              {...register("phash_distance", { valueAsNumber: true })}
            />
            <Form.Text>
              Binary distance considered a match when querying with a pHash
              fingerprint. Using more than 8 is not recommended. (Default: 0)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Draft Time Limit (seconds)</Form.Label>
            <Form.Control
              type="number"
              {...register("draft_time_limit", { valueAsNumber: true })}
            />
            <Form.Text>
              Time before a draft is deleted. (Default: 86400 = 24h)
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={4} className="mb-3">
            <Form.Label>Profiler Port</Form.Label>
            <Form.Control
              type="number"
              {...register("profiler_port", { valueAsNumber: true })}
            />
            <Form.Text>
              Port on which to serve pprof output. Omit to disable entirely.
              (Default: 0 = disabled)
            </Form.Text>
          </Form.Group>
        </Row>

        <Row>
          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>User Log File</Form.Label>
            <Form.Control
              placeholder="/path/to/user.log"
              {...register("user_log_file")}
            />
            <Form.Text>
              Path to the user log file, which logs user operations. If not set,
              these will be output to stderr.
            </Form.Text>
          </Form.Group>

          <Form.Group as={Col} md={6} className="mb-3">
            <Form.Label>Content Security Policy</Form.Label>
            <Form.Control placeholder="default-src 'self'" {...register("csp")} />
            <Form.Text>Contents of the Content-Security-Policy header.</Form.Text>
          </Form.Group>
        </Row>

        {/* Submit Button */}
        <Row>
          <Col>
            <Button type="submit" disabled={isSubmitting} className="mt-3">
              {isSubmitting ? "Saving..." : "Save Configuration"}
            </Button>
          </Col>
        </Row>
      </Form>
    </div>
  );
};

export default Settings;
