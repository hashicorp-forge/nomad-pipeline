id = "linear-regression"

job "run" {

  nomad_namespace = "default"

  resource {
    cpu    = 1000
    memory = 1024
  }

  step "install-python" {
    run = <<EOH
apt-get -q update
DEBIAN_FRONTEND=noninteractive apt-get install -y python3 python3-pip python3.12-venv
EOH
  }

  step "install-python-libs" {
    run = <<EOH
python3 -m venv regression
source regression/bin/activate
pip3 install scikit-learn numpy
EOH
  }

  step "write-script" {
    run = <<EOH
printf "
import numpy as np
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error

# Generate data
X = np.random.rand(100, 1)
y = 3 + 2 * X + np.random.randn(100, 1)

# Train model
model = LinearRegression()
model.fit(X, y)

# Evaluate model
y_pred = model.predict(X)
mse = mean_squared_error(y, y_pred)
print(f'MSE: {mse:.3f}')
" >> regression.py
EOH
  }

  step "run-script" {
    run = <<EOH
source regression/bin/activate
python3 regression.py
EOH
  }
}
