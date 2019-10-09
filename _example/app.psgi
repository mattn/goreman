#!perl

my $app = sub {
  [200, [], ["hello $ENV{AUTHOR}"]]
}
